package handler

import (
	"encoding/json"
	"log"
	"sync"

	"clean-arch-challenge-go/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderCreatedHandler struct {
	RabbitMQChannel *amqp.Channel
}

func (h *OrderCreatedHandler) Handle(event events.EventInterface, wg *sync.WaitGroup) {
	defer wg.Done()
	payload, err := json.Marshal(event.GetPayload())
	if err != nil {
		log.Fatalf("Failed to marshal event payload: %v", err)
	}
	err = h.RabbitMQChannel.Publish(
		"",
		"order_created_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	log.Printf("Event %s published to RabbitMQ", event.GetName())
}
