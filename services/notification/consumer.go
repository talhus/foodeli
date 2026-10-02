package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"foodeli/platform/events"
	platformRabbitMQ "foodeli/platform/rabbitmq"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationEvent struct {
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	OrderID   string    `json:"order_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Consumer struct {
	client    *platformRabbitMQ.Client
	channel   *amqp.Channel
	queueName string
	hub       *Hub
}

func NewConsumer(client *platformRabbitMQ.Client, hub *Hub) (*Consumer, error) {
	ch, err := client.Conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open consumer channel: %w", err)
	}

	queueName := "notifications.queue"

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("failed to declare queue '%s': %w", queueName, err)
	}
	// bind order created routing key
	err = ch.QueueBind(
		queueName,
		events.OrderCreatedRoutingKey,
		client.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("failed to bind queue to '%s': %w", events.OrderCreatedRoutingKey, err)
	}
	// bind payment succeeded routing key
	err = ch.QueueBind(
		queueName,
		events.PaymentSucceededRoutingKey,
		client.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("failed to bind queue to '%s': %w", events.PaymentSucceededRoutingKey, err)
	}
	return &Consumer{
		client:    client,
		channel:   ch,
		queueName: queueName,
		hub:       hub,
	}, nil
}

// START
func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"notification-service",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming '%s': %w", c.queueName, err)
	}
	log.Printf("[NOTIFICATION SERVICE] Listening for events on queue '%s'...", c.queueName)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[NOTIFICATION SERVICE] Context cancelled, stopping consumer...")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("[NOTIFICATION SERVICE] RabbitMQ channel closed")
					return
				}

				c.handleMessage(msg)
			}
		}
	}()

	return nil
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	defer func() {
		_ = msg.Ack(false)
	}()

	switch msg.RoutingKey {
	case events.OrderCreatedRoutingKey:
		var payload events.OrderCreatedPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			log.Printf("[NOTIFICATION SERVICE] Error unmarshaling order.created: %v", err)
			return
		}

		eventData, err := json.Marshal(NotificationEvent{
			Type:      "order.created",
			Title:     "Order Placed",
			Message:   fmt.Sprintf("Your order #%s has been placed successfully!", payload.OrderID.String()[:8]),
			OrderID:   payload.OrderID.String(),
			CreatedAt: payload.OccurredAt,
		})
		if err == nil {
			c.hub.SendToUser(payload.UserID, eventData)
			log.Printf("[Notification] Dispatched %s to user %s (order: %s)", msg.RoutingKey, payload.UserID, payload.OrderID)
		}

	case events.PaymentSucceededRoutingKey:
		var payload events.PaymentSucceededPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			log.Printf("[Notification] Error unmarshaling payment.succeeded: %v", err)
			return
		}

		eventData, err := json.Marshal(NotificationEvent{
			Type:      "payment.succeeded",
			Title:     "Payment Successful",
			Message:   fmt.Sprintf("Payment of €%.2f received for order #%s.", float64(payload.Amount)/100.0, payload.OrderID.String()[:8]),
			OrderID:   payload.OrderID.String(),
			CreatedAt: payload.OccurredAt,
		})
		if err == nil {
			c.hub.SendToUser(payload.UserID, eventData)
			log.Printf("[Notification] Dispatched %s to user %s (order: %s)", msg.RoutingKey, payload.UserID, payload.OrderID)
		}

	default:
		log.Printf("[NOTIFICATION SERVICE] Unknown routing key: %s", msg.RoutingKey)
	}
}

func (c *Consumer) Close() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
}
