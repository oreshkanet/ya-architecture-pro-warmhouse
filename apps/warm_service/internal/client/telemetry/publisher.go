package telemetry

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	// Объявляем очередь (идемпотентно)
	ch.QueueDeclare("telemetry", true, false, false, false, nil)
	return &Publisher{ch: ch}, nil
}

func (p *Publisher) Publish(moduleID string, data map[string]interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = p.ch.Publish("", "telemetry", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	return err
}
