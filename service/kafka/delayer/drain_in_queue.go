package delayer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Buddy-Git/JITScheduler-svc/model"
	"github.com/Buddy-Git/JITScheduler-svc/repository"
	"github.com/Buddy-Git/JITScheduler-svc/service/kafka"
	"go.uber.org/zap"
)

type DrainInQueue struct {
	topic           string
	manager         *kafka.QueueManager
	eventRepository repository.EventRepository
}

func NewDrainInQueue(topic string, queueManager *kafka.QueueManager, eventRepository repository.EventRepository) *DrainInQueue {
	return &DrainInQueue{
		topic:           topic,
		manager:         queueManager,
		eventRepository: eventRepository,
	}
}

func (d *DrainInQueue) DelayEvent(ctx context.Context, event *model.Event) error {
	if event.Status != model.Status.SCHEDULED {
		logger.Error("Drain In queue must get SCHEDULED events only")
	}
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
		logger.Error("Error while publishing the message:", zap.Error(err))
		return err
	}
	return nil
}

func (d *DrainInQueue) ProcessEvent(ctx context.Context, event *model.Event) error {
	logger.Info("Reached drainIn to schedule the event")
	if shouldProcess, err := ShouldProcessEvent(ctx, event, d.eventRepository); err != nil {
		logger.Error("Error:", zap.Error(err))
		return err
	} else if !shouldProcess {
		logger.Info("Ignoring the event, need to wait for the right time", zap.Any("Event", event))
		return nil
	}
	timeRemainingInSecs := event.ProcessAt.Unix() - time.Now().Unix()
	logger.Info("In drainInQueue: time remaining in secs", zap.Any("secs", timeRemainingInSecs))
	q := d.manager.GetNextQueue(timeRemainingInSecs)
	err := q.DelayEvent(ctx, event)
	if err != nil {
		err = d.eventRepository.UpdateStatus(ctx, event.ID, event.TenantID, model.Status.FAILED)
		if err != nil {
			logger.Error("Error:", zap.Error(err))
			fmt.Println("Error while moving status to failed", err)
		}
	}
	return nil
}

func (d *DrainInQueue) DelayTimeInSecs() int64 {
	return 0
}
