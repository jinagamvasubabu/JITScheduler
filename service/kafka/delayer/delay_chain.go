package delayer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"github.com/jinagamvasubabu/JITScheduler-svc/repository"
	"github.com/jinagamvasubabu/JITScheduler-svc/service/kafka"
	"go.uber.org/zap"
)

type DelayQueue struct {
	visibleTimeInSecs int64
	topic             string
	queueManager      *kafka.QueueManager
	drainOutQueue     *DrainOutQueue
	eventRepository   repository.EventRepository
}

func NewDelayQueue(visibleTimeInSecs int64, topic string, queueManager *kafka.QueueManager, drainOutQueue *DrainOutQueue, eventRepository repository.EventRepository) *DelayQueue {
	return &DelayQueue{
		visibleTimeInSecs: visibleTimeInSecs,
		topic:             topic,
		drainOutQueue:     drainOutQueue,
		eventRepository:   eventRepository,
		queueManager:      queueManager,
	}
}

func (d *DelayQueue) DelayEvent(ctx context.Context, event *model.Event) error {
	event.UpdatedAt = time.Now()
	event.CurrentQueue = d.topic
	eventPayload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Error while delay event", zap.Error(err))
		return err
	}
	logger.Info("event in delay queue", zap.Any("event", event))
	syncProducer, err := kafka.NewSyncProducer()
	if err != nil {
		return err
	}
	logger.Info("Publishing the message to next suitable topic", zap.Any("topic", d.topic), zap.Any("eventId", event.ID))
	if err := d.eventRepository.UpdateCurrentQueue(ctx, event.ID, event.TenantID, d.topic); err != nil {
		logger.Error("Error while updating current queue", zap.Error(err))
		return err
	}
	if err := syncProducer.PublishMessage(ctx, d.topic, event.ID, eventPayload); err != nil {
		logger.Error("Error while publishing event", zap.Error(err))
		return err
	}
	logger.Info("Delayed the event", zap.Any("Id", event.ID))
	return nil
}

func (d *DelayQueue) ProcessEvent(ctx context.Context, event *model.Event) error {
	if shouldProcess, err := ShouldProcessEvent(ctx, event, d.eventRepository); err != nil {
		logger.Error("Error:", zap.Error(err))
		return err
	} else if !shouldProcess {
		logger.Info("Ignored the event", zap.Any("Event", event))
		return nil
	}
	timeRemainingInSecs := event.ProcessAt.Unix() - time.Now().Unix()
	logger.Info("Remaining time to schedule the event", zap.Any("timeInSecs", timeRemainingInSecs))
	if d.visibleTimeInSecs == 0 || timeRemainingInSecs <= 0 {
		logger.Info("On time, pushing to drainOutTopic", zap.Any("event", event.ID))
		err := d.drainOutQueue.DelayEvent(ctx, event)
		if err != nil {
			err = d.eventRepository.UpdateStatus(ctx, event.ID, event.TenantID, model.Status.FAILED)
			logger.Error("error while moving event to failed status", zap.Error(err))
		}
		return nil
	}
	localTimeRemainingInSecs := d.visibleTimeInSecs - (time.Now().Unix() - event.UpdatedAt.Unix())
	if localTimeRemainingInSecs > 0 {
		logger.Info("Waiting for the time to pass the remaining secs - sleeping", zap.Any("Localtime", 5*(localTimeRemainingInSecs/60)))
		time.Sleep(time.Duration(3) * time.Second)
		timeRemainingInSecs = event.ProcessAt.Unix() - time.Now().Unix()
	}
	logger.Info("Getting next queue")
	nextQueue := d.queueManager.GetNextQueue(timeRemainingInSecs)
	err := nextQueue.DelayEvent(ctx, event)
	if err != nil {
		err := d.eventRepository.UpdateStatus(ctx, event.ID, event.TenantID, model.Status.FAILED)
		if err != nil {
			logger.Error("error while updating status of event to failed status", zap.Error(err))
		}
	}
	return nil
}

func (d *DelayQueue) DelayTimeInSecs() int64 {
	return d.visibleTimeInSecs
}
