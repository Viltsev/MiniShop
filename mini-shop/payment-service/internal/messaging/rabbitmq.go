package messaging

import (
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQ) Publish(queue string, body []byte) error {
	err := r.channel.ExchangeDeclare(
		"minishop", // exchange
		"topic",    // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	err = r.channel.Publish(
		"minishop", // exchange
		queue,      // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	return err
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
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
