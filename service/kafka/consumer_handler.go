package kafka

import (
	"encoding/json"

	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Buddy-Git/JITScheduler-svc/model"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

// ----- batch handler
type ConsumerGroupHandler interface {
	sarama.ConsumerGroupHandler
	WaitReady()
	Reset()
}

type ConsumerSessionMessage struct {
	Session sarama.ConsumerGroupSession
	Message *sarama.ConsumerMessage
}

type MultiAsyncConsumerConfig struct {
	BufChan chan *ConsumerSessionMessage
}

type multiAsyncConsumerGroupHandler struct {
	cfg *MultiAsyncConsumerConfig

	ready chan bool
}

func NewMultiAsyncConsumerGroupHandler(cfg *MultiAsyncConsumerConfig) ConsumerGroupHandler {
	handler := multiAsyncConsumerGroupHandler{
		ready: make(chan bool),
	}

	handler.cfg = cfg

	return &handler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *multiAsyncConsumerGroupHandler) Setup(s sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *multiAsyncConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *multiAsyncConsumerGroupHandler) WaitReady() {
	<-h.ready
}

func (h *multiAsyncConsumerGroupHandler) Reset() {
	h.ready = make(chan bool)
}

func (h *multiAsyncConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	claimMsgChan := claim.Messages()
	for message := range claimMsgChan {
		logger.Debug("Message Claimed Successfully!", zap.Any("message", message))
		//session.MarkMessage(message, "")
		h.cfg.BufChan <- &ConsumerSessionMessage{
			Session: session,
			Message: message,
		}
	}

	return nil
}

func decodeMessage(data []byte) error {
	var msg model.Event
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}
	return nil
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		logger.Info("Resuming consumption")
	} else {
		client.PauseAll()
		logger.Info("Pausing consumption")
	}
	*isPaused = !*isPaused
}
