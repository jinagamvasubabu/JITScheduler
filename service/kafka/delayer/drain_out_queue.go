package delayer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Buddy-Git/JITScheduler-svc/model"
	"github.com/Buddy-Git/JITScheduler-svc/repository"
	"github.com/Buddy-Git/JITScheduler-svc/service/kafka"
	"go.uber.org/zap"
)

type DrainOutQueue struct {
	eventRepository repository.EventRepository
	manager         *repository.Manager
	topic           string
}

func NewDrainOutQueue(topic string, manager *repository.Manager, eventRepository repository.EventRepository) *DrainOutQueue {
	return &DrainOutQueue{
		topic:           topic,
		eventRepository: eventRepository,
		manager:         manager,
	}
}

func (d *DrainOutQueue) DelayEvent(ctx context.Context, event *model.Event) error {
	logger.Info("In Drain out queue and the event status is =", zap.String("event status", event.Status))
	eventPayload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Error:", zap.Error(err))
		return err
	}
	syncProducer, err := kafka.NewSyncProducer()
	if err != nil {
		logger.Error("Error:", zap.Error(err))
		return err
	}
	if err := syncProducer.PublishMessage(ctx, d.topic, event.ID, eventPayload); err != nil {
		logger.Error("Error:", zap.Error(err))
		return err
	}
	return nil
}

func (d *DrainOutQueue) ProcessEvent(ctx context.Context, event *model.Event) error {
	if shouldProcess, err := ShouldProcessEvent(ctx, event, d.eventRepository); err != nil {
		return err
	} else if !shouldProcess {
		logger.Info("Ignoring the event, need to wait for the right time", zap.Any("Event", event))
		return nil
	}
	var err error
	err = d.eventRepository.UpdateStatus(ctx, event.ID, event.TenantID, model.Status.COMPLETED)
	if err != nil {
		logger.Error("Error:", zap.Error(err))
		return err
	}
	err = d.eventRepository.UpdateCurrentQueue(ctx, event.ID, event.TenantID, "drain_out_topic")
	if err != nil {
		logger.Error("Error:", zap.Error(err))
		return err
	}

	delta := time.Now().Unix() - event.ProcessAt.Unix()
	if delta > 0 {
		logger.Info("BAD: delta seconds", zap.Any("delta", delta))
	}

	if delta < 0 {
		logger.Info("BAD: shorter delay seconds", zap.Any("delta", delta))
	}

	if delta == 0 {
		logger.Info("GOOD: perfect")
	}
	err = d.manager.EventHandOff(ctx, event)
	if err != nil {
		logger.Error("error while event handing off", zap.Error(err))
	}
	return nil
}

func (d *DrainOutQueue) DelayTimeInSecs() int64 {
	return 0
}
