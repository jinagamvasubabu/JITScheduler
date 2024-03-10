package handler

import (
	"errors"
	"regexp"
	"time"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/model/dto"
	"go.uber.org/zap"
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

	istLocation, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		logger.Error("Error loading IST timezone", zap.Error(err))
		return err
	}

	/**
		package main

	import (
		"errors"
		"fmt"
		"time"
	)

	func main() {
		// Assuming date1 is provided by the client in ISO 8601 format
		date1 := "2024-03-06T23:56:00Z" // Change this to your date1 provided by the client

		// Parse date1 string into a time.Time object
		date1Time, err := time.Parse(time.RFC3339, date1)
		if err != nil {
			fmt.Println("Error parsing date1:", err)
			return
		}

		// Add 10 minutes to the current time
		targetTime := time.Now().Add(10 * time.Minute)

		// Check if date1 is greater than targetTime
		if date1Time.After(targetTime) {
			fmt.Println("date1 is greater than current time plus 10 minutes")
		} else {
			fmt.Println("date1 is not greater than current time plus 10 minutes")
		}

		// Alternatively, you can use a comparison with time.Sub() function
		if date1Time.Sub(targetTime) > 0 {
			fmt.Println("date1 is greater than current time plus 10 minutes")
		} else {
			fmt.Println("date1 is not greater than current time plus 10 minutes")
		}

		// If you need to return an error
		if date1Time.After(targetTime) {
			fmt.Println("date1 is greater than current time plus 10 minutes")
		} else {
			err := errors.New("date1 is not greater than current time plus 10 minutes")
			fmt.Println(err)
		}
	}
	*/
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
