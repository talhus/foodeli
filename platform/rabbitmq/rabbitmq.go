package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// client manages rabbitmq
type Client struct {
	Conn         *amqp.Connection
	Channel      *amqp.Channel
	ExchangeName string
}

func NewRabbitMQClient(amqpURL, exchangeName string) (*Client, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	//open a virtucal comm. channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		exchangeName,
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
		return nil, fmt.Errorf("failed to declare exgange '%s': %w", exchangeName, err)
	}

	log.Printf("Succesfully connected to RabbitMQ and declared exchange '%s'", exchangeName)

	return &Client{
		Conn:         conn,
		Channel:      ch,
		ExchangeName: exchangeName,
	}, nil
}

// publish
func (c *Client) Publish(ctx context.Context, routingKey string, body []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := c.Channel.PublishWithContext(ctx, c.ExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (c *Client) PublishJSON(ctx context.Context, routingKey string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return c.Publish(ctx, routingKey, data)
}

func (c *Client) Close() {
	if c.Channel != nil {
		_ = c.Channel.Close()
	}
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
	log.Println("RabbitMQ connection and channel closed cleanly")
}
