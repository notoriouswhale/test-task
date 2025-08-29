package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type Config struct {
	Endpoint string
	Topic    string
	GroupID  string
}

type KafkaConsumer struct {
	consumer *kafka.Consumer
	topic    string
	groupID  string

	logger *zap.Logger
	wg     sync.WaitGroup
}

func NewKafkaConsumer(cfg Config, logger *zap.Logger) (MessageBroker, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":  cfg.Endpoint,
		"group.id":           cfg.GroupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	k := &KafkaConsumer{
		consumer: consumer,
		topic:    cfg.Topic,
		groupID:  cfg.GroupID,
		logger:   logger.Named("KafkaConsumer"),
	}

	return k, nil
}

func (k *KafkaConsumer) Consume(ctx context.Context, workerCount int, handler func(message []byte) error) error {
	tasks := make(chan *kafka.Message)

	for range workerCount {
		k.wg.Add(1)
		go func() {
			defer k.wg.Done()
			for msg := range tasks {
				err := handler(msg.Value)
				if err != nil {
					k.logger.Error("Failed to process message",
						zap.Error(err),
						zap.String("message", string(msg.Value)),
					)
				}
			}
		}()
	}

	if err := k.consumer.Subscribe(k.topic, nil); err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", k.topic, err)
	}

	k.logger.Info("Started consuming messages from topic",
		zap.String("topic", k.topic),
	)

	for {
		select {
		case <-ctx.Done():
			k.logger.Info("Stopping Kafka consumer...")
			close(tasks)

			k.wg.Wait()

			return ctx.Err()

		default:
			msg, err := k.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				k.logger.Error("Consumer error",
					zap.Error(err),
				)
				continue
			}

			if msg != nil {
				select {
				case tasks <- msg:
				case <-ctx.Done():
					k.logger.Info("Stopping Kafka consumer...")
					close(tasks)
					k.wg.Wait()
					return ctx.Err()
				}
			}
		}
	}
}

func (k *KafkaConsumer) Close() error {
	k.wg.Wait()
	if k.consumer != nil {
		return k.consumer.Close()
	}
	return nil
}
