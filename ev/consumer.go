package ev

import (
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerOptions struct {
	RabbitMQUrl string
	Exchange    string
	QueueName   string
	RoutingKeys []string
}

type Consumer struct {
	RabbitMQUrl string
	Exchange    string
	QueueName   string
	channel     *amqp.Channel
}

func NewConsumer(options ConsumerOptions) (*Consumer, error) {
	conn, err := amqp.Dial(options.RabbitMQUrl)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		options.Exchange, // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		options.QueueName, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return nil, err
	}

	for _, routingKey := range options.RoutingKeys {
		err = ch.QueueBind(
			queue.Name,       // queue name
			routingKey,       // routing key
			options.Exchange, // exchange
			false,            // no-wait
			nil,              // arguments
		)
		if err != nil {
			return nil, err
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		ch.Close()
		conn.Close()
		os.Exit(1)
	}()

	return &Consumer{
		RabbitMQUrl: options.RabbitMQUrl,
		Exchange:    options.Exchange,
		QueueName:   options.QueueName,
		channel:     ch,
	}, nil
}

func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		c.QueueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
}
