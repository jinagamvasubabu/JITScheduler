package handler

import (
	"errors"
	"regexp"
	"time"

	"github.com/jinagamvasubabu/JITScheduler-svc/model/dto"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func ValidateTenant(req *dto.Tenant) error {
	if len(req.Name) == 0 {
		return errors.New("name is mandatory")
	}
	if req.Mode != "Kafka" {
		return errors.New("invalid mode")
	}
	if len(req.Properties.Brokers) == 0 {
		return errors.New("brokers needs to be specified")
	}
	if len(req.Properties.Topic) == 0 {
		return errors.New("topic needs to be specified")
	}
	if !isEmailValid(req.UpdatedBy) {
		return errors.New("incorrect format! specify gsuite id instead")
	}

	return nil
}

func ValidateEvent(req *dto.Event) error {

	if req.TenantID < 0 {
		return errors.New("invalid tenant id")
	}

	if req.Type != "DELAY" {
		return errors.New("invalid type chosen (choose DELAY)")
	}

	processAt, err := time.Parse(time.RFC3339, req.ProcessAt)
	if err != nil {
		return errors.New("incorrect format for process_at! Please follow RFC3339 format")
	}

	if processAt.Before(time.Now().Add(10 * time.Minute)) {
		return errors.New("do not create events in the past or with process_at time less than 10 mins")
	}

	if !isEmailValid(req.UpdatedBy) {
		return errors.New("incorrect format! specify gsuite id instead")
	}

	return nil
}

func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
