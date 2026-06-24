package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"

	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/judge"
)

const agentPort int32 = 8000

const (
	labelApp       = "app"
	labelLanguage  = "language"
	labelManagedBy = "managed-by"

	appExecAgent   = "exec-agent"
	managedByValue = "code-eval-service"
)

const readyTimeout = 60 * time.Second

const fillConcurrency = 4

const maxFillBackoff = 30 * time.Second

var deleteGrace = ptr.To(int64(2))

type Pod struct {
	Name string
	URL  string
}

type Pool struct {
	client       kubernetes.Interface
	namespace    string
	lang         judge.Language
	image        string
	size         int
	runtimeClass string

	ready  chan *Pod
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type Options struct {
	Client       kubernetes.Interface
	Namespace    string
	Language     judge.Language
	Image        string
	Size         int
	RuntimeClass string
}

func NewPool(opts Options) *Pool {
	return &Pool{
		client:       opts.Client,
		namespace:    opts.Namespace,
		lang:         opts.Language,
		image:        opts.Image,
		size:         opts.Size,
		runtimeClass: opts.RuntimeClass,
		ready:        make(chan *Pod, opts.Size),
	}
}

func (p *Pool) Start(ctx context.Context) error {
	if err := p.reap(ctx); err != nil {
		return fmt.Errorf("reap orphans: %w", err)
	}

	fctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	n := min(fillConcurrency, p.size)
	for range n {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.fill(fctx)
		}()
	}
	return nil
}

func (p *Pool) Shutdown(ctx context.Context) {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	if err := p.reap(ctx); err != nil {
		log.Warn().Err(err).Str("language", string(p.lang)).Msg("failed to drain exec pods")
	}
}

func (p *Pool) Ready() bool {
	return p.size == 0 || len(p.ready) > 0
}

func (p *Pool) fill(ctx context.Context) {
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}

		pod, err := p.createOne(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Warn().Err(err).Str("language", string(p.lang)).Msg("failed to warm exec pod")
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, maxFillBackoff)
			continue
		}
		backoff = time.Second

		select {
		case p.ready <- pod:
		case <-ctx.Done():
			p.deleteByName(context.Background(), pod.Name)
			return
		}
	}
}

func (p *Pool) reap(ctx context.Context) error {
	return p.client.CoreV1().Pods(p.namespace).DeleteCollection(ctx,
		metav1.DeleteOptions{GracePeriodSeconds: deleteGrace},
		metav1.ListOptions{LabelSelector: p.selector()},
	)
}

func (p *Pool) selector() string {
	return fmt.Sprintf("%s=%s,%s=%s",
		labelManagedBy, managedByValue, labelLanguage, p.lang)
}

func (p *Pool) Lease(ctx context.Context) (*Pod, error) {
	if p.size == 0 {
		return p.createOne(ctx)
	}
	select {
	case pod := <-p.ready:
		return pod, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Pool) Release(pod *Pod) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p.deleteByName(ctx, pod.Name)
}

func (p *Pool) createOne(ctx context.Context) (*Pod, error) {
	created, err := p.client.CoreV1().Pods(p.namespace).Create(ctx, p.podSpec(), metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	name := created.Name

	wctx, cancel := context.WithTimeout(ctx, readyTimeout)
	defer cancel()
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-wctx.Done():
			p.deleteByName(context.Background(), name)
			return nil, fmt.Errorf("pod %s not ready: %w", name, wctx.Err())
		case <-tick.C:
			pod, err := p.client.CoreV1().Pods(p.namespace).Get(wctx, name, metav1.GetOptions{})
			if err != nil {
				continue
			}
			if pod.Status.Phase == corev1.PodFailed {
				p.deleteByName(context.Background(), name)
				return nil, fmt.Errorf("pod %s failed to start", name)
			}
			if isReady(pod) && pod.Status.PodIP != "" {
				return &Pod{
					Name: name,
					URL:  fmt.Sprintf("http://%s:%d", pod.Status.PodIP, agentPort),
				}, nil
			}
		}
	}
}

func (p *Pool) deleteByName(ctx context.Context, name string) {
	_ = p.client.CoreV1().Pods(p.namespace).Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: deleteGrace,
	})
}

func (p *Pool) podSpec() *corev1.Pod {
	workSize := resource.MustParse("256Mi")

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("eval-%s-", p.lang),
			Labels: map[string]string{
				labelApp:       appExecAgent,
				labelLanguage:  string(p.lang),
				labelManagedBy: managedByValue,
			},
		},
		Spec: corev1.PodSpec{
			RuntimeClassName:             p.runtimeClassName(),
			RestartPolicy:                corev1.RestartPolicyNever,
			AutomountServiceAccountToken: ptr.To(false),
			EnableServiceLinks:           ptr.To(false),
			Containers: []corev1.Container{{
				Name:            appExecAgent,
				Image:           p.image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Ports:           []corev1.ContainerPort{{ContainerPort: agentPort}},
				Env: []corev1.EnvVar{
					{Name: "HOME", Value: "/tmp"},
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/healthz",
							Port: intstr.FromInt32(agentPort),
						},
					},
					InitialDelaySeconds: 1,
					PeriodSeconds:       1,
				},
				SecurityContext: &corev1.SecurityContext{
					RunAsNonRoot:             ptr.To(true),
					ReadOnlyRootFilesystem:   ptr.To(true),
					AllowPrivilegeEscalation: ptr.To(false),
					Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
					SeccompProfile:           &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
				},
				VolumeMounts: []corev1.VolumeMount{{Name: "work", MountPath: "/tmp"}},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("768Mi"),
					},
				},
			}},
			Volumes: []corev1.Volume{{
				Name: "work",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium:    corev1.StorageMediumMemory,
						SizeLimit: &workSize,
					},
				},
			}},
		},
	}
}

func (p *Pool) runtimeClassName() *string {
	if p.runtimeClass == "" {
		return nil
	}
	return &p.runtimeClass
}

func isReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}
