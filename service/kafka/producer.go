package kafka

import (
	"context"

	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

var producer Producer

type Producer interface {
	PublishMessage(ctx context.Context, topic string, key string, payload []byte) error
}

type syncProducer struct {
	instance sarama.SyncProducer
}

func NewSyncProducer() (p Producer, err error) {
	// producer config
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	// sync producer
	prd, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		logger.Error("Error while creating a new sync producer", zap.Error(err))
		return nil, err
	}

	// logger.Info("Producer instance is", zap.Any("Producer", prd))
	return &syncProducer{
		instance: prd,
	}, nil
}

func GetSyncProducer() (Producer, error) {
	if producer == nil {
		producer, err := NewSyncProducer()
		// logger.Info("Get instance is", zap.Any("Producer", producer))
		if err != nil {
			logger.Error("Error while gettign the sync producer instance", zap.Error(err))
			return producer, err
		}
	}
	return producer, nil
}

func (s syncProducer) PublishMessage(ctx context.Context, topic string, key string, payload []byte) error {
	// publish sync
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(payload),
	}
	p, o, err := s.instance.SendMessage(msg)
	if err != nil {
		logger.Error("Error while publishing a new message:", zap.Error(err))
		return err
	}
	logger.Debug("Partition: ", zap.Any("partition:", p))
	logger.Debug("Offset: ", zap.Any("offset:", o))
	return nil
}
