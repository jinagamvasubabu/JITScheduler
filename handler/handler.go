package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jinagamvasubabu/JITScheduler-svc/model"
)

func EncoderHandler(w *http.ResponseWriter, errCode int, v any) {
	//Enable CORS
	CorsHandler(w)
	(*w).Header().Add("Content-Type", "application/json")

	switch errCode {
	case 200:
		(*w).WriteHeader(http.StatusOK)
	case 201:
		(*w).WriteHeader(http.StatusCreated)
	case 400:
		(*w).WriteHeader(http.StatusBadRequest)
	case 500:
		(*w).WriteHeader(http.StatusInternalServerError)
	}
	switch v.(type) {
	case string:
		if errCode == 400 {
			response := model.ErrorResponse{
				ErrCode:    errCode,
				ErrMessage: "bad request",
				Message:    v,
			}
			json.NewEncoder(*w).Encode(response)
		} else if errCode == 500 {
			response := model.ErrorResponse{
				ErrCode:    errCode,
				ErrMessage: "internal server error",
				Message:    v,
			}
			json.NewEncoder(*w).Encode(response)
		} else {
			response := model.CorrectResponse{
				Message: v,
			}
			json.NewEncoder(*w).Encode(response)
		}
	default:
		json.NewEncoder(*w).Encode(v)
	}

}

func CorsHandler(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:4001")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS,POST,PUT")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")

}
