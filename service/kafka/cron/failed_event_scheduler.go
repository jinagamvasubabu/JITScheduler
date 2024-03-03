package cron

import (
	"context"
	"encoding/json"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	db "github.com/jinagamvasubabu/JITScheduler-svc/adapters/persistence"
	"github.com/jinagamvasubabu/JITScheduler-svc/config"
	"github.com/jinagamvasubabu/JITScheduler-svc/constants"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"github.com/jinagamvasubabu/JITScheduler-svc/repository"
	"github.com/jinagamvasubabu/JITScheduler-svc/service/kafka"
	"go.uber.org/zap"
)

type rescheduleFailedEvents struct {
	instanceId      string
	eventRepository repository.EventRepository
}

type RescheduleFailedEvents interface {
	Run()
}

func NewRescheduleFailedEvents(instanceId string, eventRepository repository.EventRepository) *rescheduleFailedEvents {
	return &rescheduleFailedEvents{
		instanceId:      instanceId,
		eventRepository: eventRepository,
	}
}

func (rs *rescheduleFailedEvents) Run() {
	logger.Debug("Fetching failed events to reschedule again")
	if ok, err := db.GetRedisClient().SetNX(constants.RescheduleCron, rs.instanceId, constants.ReschedulerCronTTL).Result(); !ok || err != nil {
		return
	}
	ctx := context.Background()
	conf := config.GetConfig()
	events, err := rs.eventRepository.FetchFailedEventsWithInXminutes(constants.SchedulerEventsWithInNextXMinutes, model.Status.FAILED)
	if err != nil {
		logger.Error("Failed to get the recent failed events", zap.Error(err))
		return
	}
	logger.Debug("failed events count:", zap.Any("count", len(events)))
	eventProducer, err := kafka.NewSyncProducer()
	if err != nil {
		logger.Error("Error while fetching the sync producer", zap.Error(err))
		return
	}
	for _, event := range events {
		event.Status = model.Status.SCHEDULED
		eventPayload, err := json.Marshal(event)
		if err != nil {
			logger.Error("Error while marshalling the event", zap.Error(err))
			continue
		}
		err = rs.eventRepository.UpdateStatus(ctx, event.ID, event.TenantID, model.Status.SCHEDULED)
		if err != nil {
			logger.Error("Error while updating the status", zap.Error(err))
			continue
		}
		if err := eventProducer.PublishMessage(ctx, conf.KafkaTopic.DrainInTopic, event.ID, eventPayload); err != nil {
			continue
		} else {
			logger.Error("Error while publishing the message", zap.Error(err))
		}
	}
}
