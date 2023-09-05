package ev

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherOptions struct {
	RabbitMQUrl  string
	Exchange     string
	ExchangeType string
}

type Publisher struct {
	RabbitMQUrl  string
	Exchange     string
	ExchangeType string
	channel      *amqp.Channel
}

type RoutingKey string

const (
	OrderCreated   RoutingKey = "order.created"
	OrderCancelled RoutingKey = "order.cancelled"
)

func NewPublisher(options PublisherOptions) (*Publisher, error) {
	conn, err := amqp.Dial(options.RabbitMQUrl)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	var args amqp.Table

	if options.ExchangeType == "x-delayed-message" {
		args = amqp.Table{
			"x-delayed-type": "direct",
		}
	}

	err = ch.ExchangeDeclare(
		options.Exchange,     // name
		options.ExchangeType, // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		args,                 // arguments
	)
	if err != nil {
		return nil, err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		ch.Close()
		conn.Close()
		os.Exit(1)
	}()

	return &Publisher{
		RabbitMQUrl:  options.RabbitMQUrl,
		Exchange:     options.Exchange,
		ExchangeType: options.ExchangeType,
		channel:      ch,
	}, nil
}

func (p *Publisher) Publish(routingKey RoutingKey, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	return p.channel.PublishWithContext(
		ctx,
		p.Exchange,         // exchange
		string(routingKey), // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}
