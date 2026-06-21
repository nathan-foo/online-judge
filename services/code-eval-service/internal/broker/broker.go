package broker

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	CodeEvalExchange    = "code_eval"
	EvalRequestsQueue   = "code_eval.requests"
	EvalResultsQueue    = "code_eval.results"
	EvalDeadLetterQueue = "code_eval.requests.dead"
)

var ErrPoison = errors.New("poison message")

type Broker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (b *Broker) Connect(url string, prefetch int) error {
	conn, err := amqp.Dial(url)

	if err != nil {
		return err
	}

	ch, err := conn.Channel()

	if err != nil {
		conn.Close()
		return err
	}

	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}

	if err := ch.ExchangeDeclare(
		CodeEvalExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	for _, q := range []string{EvalRequestsQueue, EvalResultsQueue, EvalDeadLetterQueue} {
		if _, err := ch.QueueDeclare(
			q, true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(
			q, q, CodeEvalExchange, false, nil); err != nil {
			return err
		}
	}

	b.conn, b.ch = conn, ch
	return nil
}

func (b *Broker) ConsumeRequests(handler func([]byte) error) error {
	deliveries, err := b.ch.Consume(
		EvalRequestsQueue, "", false, false, false, false, nil)

	if err != nil {
		return err
	}

	go func() {
		for d := range deliveries {
			go func(d amqp.Delivery) {
				switch err := handler(d.Body); {
				case err == nil:
					d.Ack(false)
				case errors.Is(err, ErrPoison):
					b.deadLetter(d)
				case d.Redelivered:
					b.deadLetter(d)
				default:
					log.Warn().Err(err).Msg("requeueing eval request after transient failure")
					d.Nack(false, true)
				}
			}(d)
		}
	}()

	return nil
}

func (b *Broker) deadLetter(d amqp.Delivery) {
	if err := b.PublishDeadLetter(context.Background(), d.Body); err != nil {
		log.Error().Err(err).Msg("failed to dead-letter eval request; requeueing")
		d.Nack(false, true)
		return
	}
	log.Warn().Msg("dead-lettered eval request")
	d.Ack(false)
}

func (b *Broker) PublishResult(ctx context.Context, body []byte) error {
	return b.publish(ctx, EvalResultsQueue, body)
}

func (b *Broker) PublishDeadLetter(ctx context.Context, body []byte) error {
	return b.publish(ctx, EvalDeadLetterQueue, body)
}

func (b *Broker) publish(ctx context.Context, routingKey string, body []byte) error {
	return b.ch.PublishWithContext(ctx, CodeEvalExchange, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
}

func (b *Broker) Close() error {
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *Broker) Healthy() error {
	if b.conn == nil || b.conn.IsClosed() {
		return errors.New("broker connection is closed")
	}
	return nil
}
