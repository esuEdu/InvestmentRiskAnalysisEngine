package consumer

import (
	"context"

	rabbitmqamqp "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type Consumer struct {
	conn *rabbitmqamqp.AmqpConnection
}

func NewConsumer(conn *rabbitmqamqp.AmqpConnection) *Consumer {
	return &Consumer{conn: conn}
}

func (c *Consumer) Consume(ctx context.Context, queue string, handler func([]byte) error) error {
	_, err := c.conn.Management().DeclareQueue(ctx, &rabbitmqamqp.ClassicQueueSpecification{
		Name: queue,
	})
	if err != nil {
		return err
	}

	cons, err := c.conn.NewConsumer(ctx, queue, nil)
	if err != nil {
		return err
	}
	defer cons.Close(ctx)

	for {
		delivery, err := cons.Receive(ctx)
		if err != nil {
			return err
		}

		var body []byte
		if delivery.Message().GetData() != nil {
			body = delivery.Message().GetData()
		}

		if err := handler(body); err != nil {
			if requeueErr := delivery.Requeue(ctx); requeueErr != nil {
				return requeueErr
			}
			continue
		}

		if err := delivery.Accept(ctx); err != nil {
			return err
		}
	}
}
