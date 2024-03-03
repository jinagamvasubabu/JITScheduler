package kafka

import (
	"context"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"go.uber.org/zap"
)

type Queue interface {
	ProcessEvent(ctx context.Context, event *model.Event) error
	DelayEvent(ctx context.Context, event *model.Event) error
	DelayTimeInSecs() int64
}

type QueueManager struct {
	DelayQueues []Queue
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		DelayQueues: make([]Queue, 0)}
}

func (qm *QueueManager) GetNextQueue(timeInSecs int64) Queue {
	logger.Info("Delay topics length", zap.Any("len", len(qm.DelayQueues)))
	index := 1
	length := len(qm.DelayQueues)
	for index < length && qm.DelayQueues[index].DelayTimeInSecs() <= timeInSecs {
		index++
	}
	logger.Info("Possible next topic is", zap.Any("next queue is:", index-1))
	return qm.DelayQueues[index-1]
}
