package delayer

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"github.com/jinagamvasubabu/JITScheduler-svc/repository"
	"go.uber.org/zap"
)

func ShouldProcessEvent(ctx context.Context, event *model.Event, eventRepository repository.EventRepository) (bool, error) {
	eventDB, err := eventRepository.FetchEvent(ctx, event.ID, event.TenantID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("Error:", zap.Error(err))
		return false, err
	}
	return eventDB != nil &&
		(eventDB.Status == model.Status.REQUESTED || eventDB.Status == model.Status.SCHEDULED) &&
		eventDB.UpdatedBy == event.UpdatedBy, nil
}
