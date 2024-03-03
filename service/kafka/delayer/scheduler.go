package delayer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/config"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"github.com/jinagamvasubabu/JITScheduler-svc/repository"
	"github.com/jinagamvasubabu/JITScheduler-svc/service/kafka"
	"go.uber.org/zap"
)

type Scheduler struct {
	stopChan         chan bool
	events           map[string]*model.Event
	queueManager     *kafka.QueueManager
	manager          *repository.Manager
	eventRepository  repository.EventRepository
	tenantRepository repository.TenantRepository
}

func NewScheduler(ctx context.Context, eventRepository repository.EventRepository, tenantRepository repository.TenantRepository) Scheduler {
	return Scheduler{
		stopChan:         make(chan bool),
		events:           make(map[string]*model.Event),
		queueManager:     kafka.NewQueueManager(),
		manager:          repository.NewManager(tenantRepository, eventRepository),
		eventRepository:  eventRepository,
		tenantRepository: tenantRepository,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigChan:
				s.stopChan <- true
			case <-s.stopChan:
				close(s.stopChan)
			}
		}
	}()
	if err := s.initkafkaQueues(ctx); err != nil {
		logger.Error("Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *Scheduler) Wait() {
	<-s.stopChan
}

func (s *Scheduler) initkafkaQueues(ctx context.Context) error {
	logger.Info("initialising Kafka Queues")
	var queueEvents []*kafka.QueueEvent
	cfg := config.GetConfig()
	//Init Drain In consumer
	drainInQueue := NewDrainInQueue(cfg.KafkaTopic.DrainInTopic, s.queueManager, s.eventRepository)
	drainInQueueEvent := kafka.NewQueueEvent(drainInQueue, cfg.KafkaTopic.DrainInTopic)
	queueEvents = append(queueEvents, drainInQueueEvent)

	//Init Drain out consumer
	drainOutDelayer := NewDrainOutQueue(cfg.KafkaTopic.DrainOutTopic, s.manager, s.eventRepository)
	drainOutDelayerQueueEvent := kafka.NewQueueEvent(drainOutDelayer, cfg.KafkaTopic.DrainOutTopic)
	queueEvents = append(queueEvents, drainOutDelayerQueueEvent)
	s.queueManager.DelayQueues = append(s.queueManager.DelayQueues, drainOutDelayer)

	//1 Second delayer
	oneSecDelayer := NewDelayQueue(1, cfg.KafkaTopic.OneSecQueueTopic, s.queueManager, drainOutDelayer, s.eventRepository)
	oneSecDelayerQueueEvent := kafka.NewQueueEvent(oneSecDelayer, cfg.KafkaTopic.OneSecQueueTopic)
	queueEvents = append(queueEvents, oneSecDelayerQueueEvent)
	s.queueManager.DelayQueues = append(s.queueManager.DelayQueues, oneSecDelayer)

	//5 Second delayer
	fiveSecDelayer := NewDelayQueue(5, cfg.KafkaTopic.FiveSecQueueTopic, s.queueManager, drainOutDelayer, s.eventRepository)
	fiveSecDelayerQueueEvent := kafka.NewQueueEvent(fiveSecDelayer, cfg.KafkaTopic.FiveSecQueueTopic)
	queueEvents = append(queueEvents, fiveSecDelayerQueueEvent)
	s.queueManager.DelayQueues = append(s.queueManager.DelayQueues, fiveSecDelayer)

	//15 Second delayer
	fifteenSecDelayer := NewDelayQueue(15, cfg.KafkaTopic.FifteenSecQueueTopic, s.queueManager, drainOutDelayer, s.eventRepository)
	fifteenSecDelayerQueueEvent := kafka.NewQueueEvent(fifteenSecDelayer, cfg.KafkaTopic.FifteenSecQueueTopic)
	queueEvents = append(queueEvents, fifteenSecDelayerQueueEvent)
	s.queueManager.DelayQueues = append(s.queueManager.DelayQueues, fifteenSecDelayer)

	//1 Min delayer
	oneMinDelayer := NewDelayQueue(60, cfg.KafkaTopic.OneMinQueueTopic, s.queueManager, drainOutDelayer, s.eventRepository)
	oneMinDelayerQueueEvent := kafka.NewQueueEvent(oneMinDelayer, cfg.KafkaTopic.OneMinQueueTopic)
	queueEvents = append(queueEvents, oneMinDelayerQueueEvent)
	s.queueManager.DelayQueues = append(s.queueManager.DelayQueues, oneMinDelayer)

	//5 Min delayer
	fiveMinDelayer := NewDelayQueue(300, cfg.KafkaTopic.FiveMinQueueTopic, s.queueManager, drainOutDelayer, s.eventRepository)
	fiveMinDelayerQueueEvent := kafka.NewQueueEvent(oneMinDelayer, cfg.KafkaTopic.FiveMinQueueTopic)
	queueEvents = append(queueEvents, fiveMinDelayerQueueEvent)
	s.queueManager.DelayQueues = append(s.queueManager.DelayQueues, fiveMinDelayer)

	//15 Min delayer
	fifteenMinDelayer := NewDelayQueue(900, cfg.KafkaTopic.FifteenMinQueueTopic, s.queueManager, drainOutDelayer, s.eventRepository)
	fifteenMinDelayerQueueEvent := kafka.NewQueueEvent(fifteenMinDelayer, cfg.KafkaTopic.FifteenMinQueueTopic)
	queueEvents = append(queueEvents, fifteenMinDelayerQueueEvent)
	s.queueManager.DelayQueues = append(s.queueManager.DelayQueues, fifteenMinDelayer)
	go func() {
		logger.Info("Creating Consumer Goup")
		kafka.NewEventConsumer(queueEvents).Consume()
	}()

	return nil
}

func (s *Scheduler) ScheduleEvent(ctx context.Context, event *model.Event) error {
	eventPayload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Error", zap.Error(err))
		return err
	}
	shouldSchedule := event.ProcessAt.Add(30 * time.Minute).After(event.ProcessAt)
	if shouldSchedule {
		logger.Info("Event is scheduled", zap.Any("event", event))
		event.Status = model.Status.SCHEDULED
	}
	if err := s.eventRepository.AddEvent(ctx, event); err != nil {
		logger.Error("Error", zap.Error(err))
		return err
	}
	if shouldSchedule {
		syncProducer, err := kafka.NewSyncProducer()
		if err != nil {
			logger.Error("Error", zap.Error(err))
			return err
		}
		if err := syncProducer.PublishMessage(ctx, config.GetConfig().KafkaTopic.DrainInTopic, event.ID, eventPayload); err != nil {
			logger.Error("Error publishing", zap.Error(err))
		}
	}
	return nil
}

func (s *Scheduler) CancelEvent(ctx context.Context, eventId string, TenantID int32) error {
	eventDB, err := s.eventRepository.FetchEvent(ctx, eventId, TenantID)
	if err != nil {
		logger.Error("Error", zap.Error(err))
		return err
	}
	if eventDB.Status == model.Status.COMPLETED {
		logger.Error("event status is already in completed state")
		return errors.New("event is already in completed state")
	}

	if err := s.eventRepository.UpdateStatus(ctx, eventId, TenantID, model.Status.CANCELLED); err != nil {
		logger.Error("Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *Scheduler) GetUnProcessedEventsCount() ([]*model.UnProcessedEventsCount, error) {
	return s.eventRepository.GetUnProcessedEventsCount()
}

func (s *Scheduler) GetUnProcessedEvents(ctx context.Context) (string, error) {
	events, err := s.eventRepository.GetUnProcessedEvents()
	if err != nil {
		logger.Error("Error", zap.Error(err))
		return "", err
	}
	payload, err := json.Marshal(events)
	if err != nil {
		logger.Error("Error", zap.Error(err))
		return "", err
	}
	return string(payload), nil
}
