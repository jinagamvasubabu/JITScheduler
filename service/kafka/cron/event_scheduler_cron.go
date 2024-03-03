package cron

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	db "github.com/jinagamvasubabu/JITScheduler-svc/adapters/persistence"
	"github.com/jinagamvasubabu/JITScheduler-svc/config"
	"github.com/jinagamvasubabu/JITScheduler-svc/constants"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"github.com/jinagamvasubabu/JITScheduler-svc/repository"
	"github.com/jinagamvasubabu/JITScheduler-svc/service/kafka"
	"go.uber.org/zap"
)

type scheduleEvents struct {
	instanceId      string
	eventRepository repository.EventRepository
}

type ScheduleEvents interface {
	Run()
}

func NewScheduleEventsCron(instanceId string, eventRepository repository.EventRepository) *scheduleEvents {
	return &scheduleEvents{
		instanceId:      instanceId,
		eventRepository: eventRepository,
	}
}

func (sc *scheduleEvents) Run() {
	if _, err := db.GetRedisClient().SetNX(constants.SchedulerCron, sc.instanceId, constants.SchedulerCronTTL).Result(); err != nil {
		logger.Error("Error while setting the lock", zap.Error(err))
		return
	}
	ctx := context.Background()
	conf := config.GetConfig()
	events, err := sc.eventRepository.FetchEventsWithInNextXmins(int32(constants.SchedulerEventsWithInNextXMinutes), model.Status.REQUESTED)
	if err != nil {
		logger.Error("Error while fetching next set of events for to schedule", zap.Error(err))
		return
	}
	if len(events) > 0 {
		logger.Info("events that can be scheduled", zap.Any("events", events))
		eventProducer, err := kafka.NewSyncProducer()
		if eventProducer == nil {
			logger.Error("Error while fetching event producer", zap.Any("error", err))
		}
		if err != nil {
			logger.Error("Error while getting the sync producer", zap.Error(err))
			return
		}
		for _, event := range events {
			if event.Status == model.Status.REQUESTED {
				event.Status = model.Status.SCHEDULED
				eventPayload, err := json.Marshal(event)
				if err != nil {
					logger.Error("Error while marshalling the event", zap.Error(err))
					return
				}
				err = sc.eventRepository.UpdateStatus(ctx, event.ID, event.TenantID, model.Status.SCHEDULED)
				if err != nil {
					logger.Error("Error while updating the status to scheduled", zap.Error(err))
					continue
				}
				if err := eventProducer.PublishMessage(ctx, conf.KafkaTopic.DrainInTopic, event.ID, eventPayload); err != nil {
					logger.Error("Error while publishing to drain in topic", zap.Error(err))
					err := sc.eventRepository.UpdateStatus(ctx, event.ID, event.TenantID, model.Status.FAILED)
					if err != nil {
						logger.Error("Error while updating the status if the event to failed", zap.Error(err))
						continue
					}
					continue
				} else {
					logger.Info(fmt.Sprintf("Successfully published the event %s to drainInTopic %s", event.ID, conf.KafkaTopic.DrainInTopic))
				}
			}
		}
	}

}
