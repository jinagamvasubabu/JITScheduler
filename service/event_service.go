package service

import (
	"context"
	"time"

	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Buddy-Git/JITScheduler-svc/model"
	"github.com/Buddy-Git/JITScheduler-svc/model/dto"
	"github.com/Buddy-Git/JITScheduler-svc/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type eventService struct {
	eventRepository repository.EventRepository
}

type EventService interface {
	AddEvent(ctx context.Context, request *dto.Event) (string, error)
	FetchAllEventsUnderTenant(ctx context.Context, TenantID int32) ([]*dto.Event, error)
	DeactivateEvent(ctx context.Context, id string, TenantID int32, user string) (string, error)
}

func NewEventService(ctx context.Context, eventRepository repository.EventRepository) EventService {
	return eventService{
		eventRepository: eventRepository,
	}
}

func (e eventService) AddEvent(ctx context.Context, request *dto.Event) (string, error) {
	event := &model.Event{}
	event.ID = uuid.New().String()
	event.TenantID = request.TenantID
	ProcessAt, _ := time.Parse(time.RFC3339, request.ProcessAt)
	event.ProcessAt = ProcessAt
	event.UpdatedAt = time.Now()
	event.UpdatedBy = request.UpdatedBy
	event.Status = model.Status.REQUESTED
	if err := e.eventRepository.AddEvent(ctx, event); err != nil {
		logger.Error("Error while creating the event = %s", zap.Error(err))
		return "", err
	}
	return "success adding an event", nil
}

func (e eventService) FetchAllEventsUnderTenant(ctx context.Context, TenantID int32) ([]*dto.Event, error) {
	var events []*model.Event
	var err error
	events, err = e.eventRepository.FetchAllEventsUnderTenant(ctx, TenantID)
	if err != nil {
		logger.Error("Error while fetching all the events", zap.Error(err))
		return nil, err
	}

	var eventsDto []*dto.Event
	for _, event := range events {
		eventsDto = append(eventsDto, &dto.Event{
			ID:        event.ID,
			TenantID:  event.TenantID,
			Type:      event.Type,
			ProcessAt: event.ProcessAt.Format(time.RFC3339),
			Status:    event.Status,
			UpdatedBy: event.UpdatedBy,
			UpdatedAt: event.UpdatedAt,
		})
	}
	return eventsDto, err

}

func (e eventService) DeactivateEvent(ctx context.Context, id string, TenantID int32, user string) (string, error) {
	if err := e.eventRepository.DeactivateEvent(ctx, id, TenantID, user); err != nil {
		logger.Error("Error while deactivating the event", zap.Error(err))
		return "", err
	}
	return "success deactivating an event", nil
}
