package broker

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	CodeEvalExchange  = "code_eval"
	EvalRequestsQueue = "code_eval.requests"
	EvalResultsQueue  = "code_eval.results"
)

type Broker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (b *Broker) Connect(url string) error {
	conn, err := amqp.Dial(url)

	if err != nil {
		return err
	}

	ch, err := conn.Channel()

	if err != nil {
		conn.Close()
		return err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	if err := ch.ExchangeDeclare(
		CodeEvalExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	for _, q := range []string{EvalRequestsQueue, EvalResultsQueue} {
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
			if err := handler(d.Body); err != nil {
				d.Nack(false, false)
				continue
			}
			d.Ack(false)
		}
	}()

	return nil
}

func (b *Broker) PublishResult(ctx context.Context, body []byte) error {
	return b.ch.PublishWithContext(ctx, CodeEvalExchange, EvalResultsQueue, false, false,
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
