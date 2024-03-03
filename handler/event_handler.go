package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"fmt"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Buddy-Git/JITScheduler-svc/model"
	"github.com/Buddy-Git/JITScheduler-svc/model/dto"
	"github.com/Buddy-Git/JITScheduler-svc/service"
)

type eventHandler struct {
	eventService service.EventService
}

func NewEventHandler(eventService service.EventService) eventHandler {
	return eventHandler{eventService}
}

func (h eventHandler) AddEvent(w http.ResponseWriter, r *http.Request) {
	// Read to request body
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, err.Error())
		return
	}

	var event dto.Event
	err = json.Unmarshal(body, &event)
	if err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, err.Error())
		return
	}
	logger.Info("event request", zap.Any("event", event))
	//VALIDATE BEFORE SENDING TO SERVICE
	if err := ValidateEvent(&event); err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, err.Error())
		return
	}

	// Append to the Event table
	if result, err := h.eventService.AddEvent(context.Background(), &event); err != nil {
		// Send a 500 bad_request response
		EncoderHandler(&w, 500, err.Error())
		return
	} else {
		// Send a 201 created response
		EncoderHandler(&w, 201, result)
		logger.Info("Event created")
		return
	}
}

func (h eventHandler) FetchAllEventsUnderTenant(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	TenantID := vars["tenant_id"]
	ID, _ := strconv.ParseInt(TenantID, 10, 32)

	var events []*dto.Event
	var err error
	if events, err = h.eventService.FetchAllEventsUnderTenant(context.Background(), int32(ID)); err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 500, err.Error())
		return
	} else {
		EncoderHandler(&w, 200, events)
		logger.Info("Events fetched")
		return
	}
}

//CHANGE THIS TO DEACTIVATE EVENTS
func (h eventHandler) DeactivateEvent(w http.ResponseWriter, r *http.Request) {
	// Read dynamic id parameter
	vars := mux.Vars(r)
	eventID := vars["event_id"]
	TenantID := vars["tenant_id"]
	ID, _ := strconv.ParseInt(TenantID, 10, 32)

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, fmt.Sprintf("Request is invalid = %s", err.Error()))
		return
	}
	var user model.User
	json.Unmarshal(body, &user)

	if result, err := h.eventService.DeactivateEvent(context.Background(), eventID, int32(ID), user.Email); err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, err.Error())
		return
	} else {
		EncoderHandler(&w, 200, result)
		logger.Info("Event deactivated")
		return
	}

}
