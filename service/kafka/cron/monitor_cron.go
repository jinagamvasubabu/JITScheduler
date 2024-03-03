package cron

import (
	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Buddy-Git/JITScheduler-svc/adapters/persistence"
	"github.com/Buddy-Git/JITScheduler-svc/constants"
	"github.com/Buddy-Git/JITScheduler-svc/repository"
	"go.uber.org/zap"
)

type monitorCron struct {
	eventRepository repository.EventRepository
}

func NewMonitorCron(eventRepository repository.EventRepository) monitorCron {
	return monitorCron{
		eventRepository: eventRepository,
	}
}

func (mr *monitorCron) MonitorQueuedEvents() {
	logger.Debug("Acquiring the lock to make sure only one instance is taking care of scheduler")
	//TODO: fix the error
	if ok, err := persistence.GetRedisClient().SetNX(constants.MonitorCron, "scheduler_monitor_cron_lock", constants.MonitorCronTTL).Result(); !ok || err != nil {
		logger.Error("Error while acquiring the monitor lock %s", zap.Any("error", err))
		return
	}
	queuedEvents, err := mr.eventRepository.GetQueuedEventsCount()
	if err != nil {
		logger.Error("Error while fetching the next queued events", zap.Error(err))
		return
	}
	logger.Info("Length of queued events are", zap.Any("quequedEvents", len(queuedEvents)))
}
