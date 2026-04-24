package publisher

import (
	"context"

	rabbitmqamqp "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type Publisher struct {
	conn *rabbitmqamqp.AmqpConnection
}

func NewPublisher(conn *rabbitmqamqp.AmqpConnection) *Publisher {
	return &Publisher{conn: conn}
}

func (p *Publisher) Publish(ctx context.Context, queue string, body []byte) error {
	_, err := p.conn.Management().DeclareQueue(ctx, &rabbitmqamqp.ClassicQueueSpecification{
		Name: queue,
	})
	if err != nil {
		return err
	}

	pub, err := p.conn.NewPublisher(ctx, &rabbitmqamqp.QueueAddress{Queue: queue}, nil)
	if err != nil {
		return err
	}
	defer pub.Close(ctx)

	_, err = pub.Publish(ctx, rabbitmqamqp.NewMessage(body))
	return err
}
