package kafka

// SIGUSR1 toggle the pause/resume consumption
import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jinagamvasubabu/JITScheduler-svc/config"

	"github.com/Shopify/sarama"
	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"go.uber.org/zap"
)

// Sarama configuration options
var (
	group = "jit_scheduler_group"
)

type EventConsumer struct {
	QueueEvents []*QueueEvent
}

func NewEventConsumer(queueEvents []*QueueEvent) *EventConsumer {
	return &EventConsumer{
		QueueEvents: queueEvents,
	}
}

func (e *EventConsumer) Consume() error {
	logger.Info("Starting a new Sarama consumer group")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	var bufChan = make(chan *ConsumerSessionMessage, 1000)

	go func() {
		for message := range bufChan {
			message.Session.MarkMessage(message.Message, "")
			topic := message.Message.Topic
			for idx, e := range e.QueueEvents {
				if e == nil {
					continue
				}
				if e.Topic == topic {
					logger.Info("match", zap.Any("Topic", topic))
					_, err := e.Call(*message.Message, idx)
					if err != nil {
						return
					}
				}
			}

		}
	}()

	handler := NewMultiAsyncConsumerGroupHandler(&MultiAsyncConsumerConfig{
		BufChan: bufChan,
	})
	brokerUrl := config.GetConfig().KafkaBrokerUrl
	client, err := sarama.NewConsumerGroup([]string{brokerUrl}, group, cfg)
	if err != nil {
		return err
	}

	keepRunning := true
	consumptionIsPaused := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	//list of topics
	var topics []string
	for _, e := range e.QueueEvents {
		if e == nil {
			continue
		}
		topics = append(topics, e.Topic)
	}

	go func() {
		for {
			logger.Info("reading----->", zap.Any("asd", topics))
			err := client.Consume(ctx, topics, handler)
			if err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					break
				} else {
					panic(err)
				}
			}

			if ctx.Err() != nil {
				return
			}
			handler.Reset()
		}
	}()

	handler.WaitReady() // Await till the consumer has been set up
	logger.Info("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	for keepRunning {
		select {
		case <-ctx.Done():
			logger.Info("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			logger.Info("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(client, &consumptionIsPaused)
		}
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		return err
	}
	return nil
}
