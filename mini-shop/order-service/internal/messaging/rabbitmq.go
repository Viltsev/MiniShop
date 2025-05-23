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

func NewRabbitMQ(amqpURL, exchange string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(amqpURL)
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

func (r *RabbitMQ) Consume(bindingKey string, handler func([]byte)) error {
	log.Printf("[RabbitMQ] Declaring exchange '%s' for binding key: %s", r.exchange, bindingKey)

	// Создаем уникальную очередь с рандомным именем (server-named queue)
	q, err := r.channel.QueueDeclare(
		"",    // пустая строка - сервер сгенерирует имя
		true,  // durable
		false, // delete when unused
		true,  // exclusive (очередь удалится, когда коннект закроется)
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("[RabbitMQ] Binding queue '%s' to exchange '%s' with routing key '%s'", q.Name, r.exchange, bindingKey)
	err = r.channel.QueueBind(
		q.Name,
		bindingKey,
		r.exchange,
		false,
		nil,
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
			log.Printf("[RabbitMQ] Message received on queue %s: %s", q.Name, string(msg.Body))
			handler(msg.Body)
		}
	}()

	log.Printf("[RabbitMQ] Waiting for messages on queue: %s", q.Name)
	return nil
}

func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}
