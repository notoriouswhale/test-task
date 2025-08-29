package messaging

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type KafkaBroker struct {
	producer *kafka.Producer
	topic    string
	config   *kafka.ConfigMap
	logger   *zap.Logger
}

type Config struct {
	Endpoint     string
	BaseClientID string
	Topic        string
}

func NewKafkaBroker(cfg Config, logger *zap.Logger) (*KafkaBroker, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("kafka endpoint is required")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	id := cfg.BaseClientID
	if h, err := os.Hostname(); err == nil {
		id = fmt.Sprintf("%s-%s", cfg.BaseClientID, h)
	}

	config := &kafka.ConfigMap{
		"bootstrap.servers":  cfg.Endpoint,
		"client.id":          id,
		"compression.codec":  "gzip",
		"acks":               "all",
		"enable.idempotence": true,
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	broker := &KafkaBroker{
		producer: producer,
		topic:    cfg.Topic,
		config:   config,
		logger:   logger.Named("KafkaBroker"),
	}

	go broker.handleDeliveryReports()

	return broker, nil
}

func (b *KafkaBroker) Send(ctx context.Context, topic string, message, key []byte) error {
	if b.producer == nil {
		return fmt.Errorf("kafka producer is not initialized")
	}

	targetTopic := b.topic
	if topic != "" {
		targetTopic = topic
	}

	kafkaMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &targetTopic,
			Partition: kafka.PartitionAny,
		},
		Value:     message,
		Key:       key,
		Timestamp: time.Now(),
	}

	deliveryChan := make(chan kafka.Event, 1)

	err := b.producer.Produce(kafkaMessage, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

func (b *KafkaBroker) handleDeliveryReports() {
	if b.producer == nil {
		return
	}

	for e := range b.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				b.logger.Error("Failed to deliver message:", zap.Error(ev.TopicPartition.Error))
			}
		case kafka.Error:
			b.logger.Error("Kafka error:", zap.Error(ev))
		}
	}
}

func (b *KafkaBroker) Close() error {
	if b.producer != nil {
		b.producer.Flush(5000)
		b.producer.Close()
	}
	return nil
}
