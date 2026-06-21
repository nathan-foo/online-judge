package pool

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"

	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/judge"
)

type Manager struct {
	pools map[judge.Language]*Pool
}

type ManagerOptions struct {
	Client       kubernetes.Interface
	Namespace    string
	Images       map[judge.Language]string
	Sizes        map[judge.Language]int
	RuntimeClass string
}

func NewManager(opts ManagerOptions) *Manager {
	pools := make(map[judge.Language]*Pool, len(opts.Images))
	for lang, image := range opts.Images {
		pools[lang] = NewPool(Options{
			Client:       opts.Client,
			Namespace:    opts.Namespace,
			Language:     lang,
			Image:        image,
			Size:         opts.Sizes[lang],
			RuntimeClass: opts.RuntimeClass,
		})
	}
	return &Manager{pools: pools}
}

func (m *Manager) Start(ctx context.Context) error {
	for lang, p := range m.pools {
		if err := p.Start(ctx); err != nil {
			return fmt.Errorf("start %s pool: %w", lang, err)
		}
	}
	return nil
}

func (m *Manager) Shutdown(ctx context.Context) {
	var wg sync.WaitGroup
	for _, p := range m.pools {
		wg.Add(1)
		go func(p *Pool) {
			defer wg.Done()
			p.Shutdown(ctx)
		}(p)
	}
	wg.Wait()
}

func (m *Manager) Lease(ctx context.Context, lang judge.Language) (*Pod, error) {
	p, ok := m.pools[lang]
	if !ok {
		return nil, fmt.Errorf("no pool for language %q", lang)
	}
	return p.Lease(ctx)
}

func (m *Manager) Release(lang judge.Language, pod *Pod) {
	if p, ok := m.pools[lang]; ok {
		p.Release(pod)
	}
}

func (m *Manager) Ready() error {
	for lang, p := range m.pools {
		if !p.Ready() {
			return fmt.Errorf("no warm pods for %s", lang)
		}
	}
	return nil
}
