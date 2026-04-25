package messaging

import (
	"context"
	"fmt"

	rabbitmqamqp "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type RabbitMQ struct {
	Conn *rabbitmqamqp.AmqpConnection
}

func NewRabbitMQ(ctx context.Context, host string, port int, user, password string) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", user, password, host, port)

	conn, err := rabbitmqamqp.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{Conn: conn}, nil
}

func (r *RabbitMQ) Close(ctx context.Context) error {
	return r.Conn.Close(ctx)
}
