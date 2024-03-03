package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"go.uber.org/zap"
)

type QueueEvent struct {
	EventHandler Queue
	Topic        string
}

func NewQueueEvent(eventHandler Queue, topic string) *QueueEvent {
	return &QueueEvent{
		EventHandler: eventHandler,
		Topic:        topic,
	}
}

func (q *QueueEvent) Call(msg sarama.ConsumerMessage, _ int) (retryAfter time.Duration, err error) {
	logger.Info("Call: msg", zap.Any("Topic", msg.Topic))
	var event model.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		logger.Error("Error:", zap.Error(err))
		return 0, err
	}

	err = q.EventHandler.ProcessEvent(context.Background(), &event)
	return 0, err
}
