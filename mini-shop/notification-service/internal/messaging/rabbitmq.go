package messaging

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

func NewRabbitMQ(ampqURL, exchange string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(ampqURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = ch.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
	}, nil
}

func (r *RabbitMQ) Publish(routingKey string, body []byte) error {
	log.Printf("Publishing message to exchange %s with routing key %s: %s", r.exchange, routingKey, string(body))
	return r.channel.Publish(
		r.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

func (r *RabbitMQ) Consume(queue string, handler func([]byte)) error {
	log.Printf("[RabbitMQ] Declaring exchange 'minishop' for queue: %s", queue)
	err := r.channel.ExchangeDeclare(
		"minishop", "topic", true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	log.Printf("[RabbitMQ] Declaring queue: %s", queue)
	q, err := r.channel.QueueDeclare(
		queue, true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	log.Printf("[RabbitMQ] Binding queue '%s' to exchange 'minishop' with routing key '%s'", q.Name, queue)
	err = r.channel.QueueBind(
		q.Name, queue, "minishop", false, nil,
	)
	if err != nil {
		return err
	}

	log.Printf("[RabbitMQ] Subscribing to queue: %s", q.Name)
	msgs, err := r.channel.Consume(
		q.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			log.Printf("[RabbitMQ] Message received on queue %s: %s", queue, string(msg.Body))
			handler(msg.Body)
		}
	}()

	log.Printf("[RabbitMQ] Waiting for messages on queue: %s", queue)
	return nil
}

func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}
