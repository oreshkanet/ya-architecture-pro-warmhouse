package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"telemetry-service/internal/domain"
	"telemetry-service/internal/service/aggregator"
	"telemetry-service/internal/service/validator"

	"github.com/streadway/amqp"
)

type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	engine  *aggregator.AggregationEngine
}

func NewRabbitMQConsumer(url, queue string, engine *aggregator.AggregationEngine) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		queue:   queue,
		engine:  engine,
	}, nil
}

func (c *RabbitMQConsumer) Start() {
	msgs, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	for msg := range msgs {
		ctx := context.Background()

		var raw map[string]interface{}
		if err := json.Unmarshal(msg.Body, &raw); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			msg.Nack(false, false)
			continue
		}

		point := domain.TelemetryPoint{
			SensorID:  raw["sensor_id"].(string),
			DeviceID:  raw["device_id"].(string),
			Type:      raw["type"].(string),
			Unit:      raw["unit"].(string),
			Value:     raw["temperature"], // TODO: обобщить под разные типы
			Timestamp: time.Now().UTC(),
		}

		if err := validator.Validate(&point); err != nil {
			log.Printf("Validation failed: %v", err)
			msg.Nack(false, false)
			continue
		}

		// Сохраняем по одной записи (в реальности — батчинг)
		if err := c.engine.ProcessBatch(ctx, []domain.TelemetryPoint{point}); err != nil {
			log.Printf("Failed to process: %v", err)
			msg.Nack(false, false)
			continue
		}

		msg.Ack(false)
	}
}

func (c *RabbitMQConsumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
