package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"github.com/jinagamvasubabu/JITScheduler-svc/model/dto"
	"github.com/jinagamvasubabu/JITScheduler-svc/service"
	"go.uber.org/zap"
)

type tenantHandler struct {
	tenantService service.TenantService
}

func NewTenantHandler(tenantService service.TenantService) tenantHandler {
	return tenantHandler{tenantService}
}

func (h tenantHandler) AddTenant(w http.ResponseWriter, r *http.Request) {
	// Read to request body
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, err.Error())
		return
	}
	//create the dto object
	var tenant dto.Tenant
	err = json.Unmarshal(body, &tenant)
	if err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, err.Error())
		return
	}
	//VALIDATE BEFORE SENDING TO SERVICE
	if err = ValidateTenant(&tenant); err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, err.Error())
		return
	}

	logger.Info("zap", zap.Any("tenant", tenant))
	// Append to the Tenant table
	if result, err := h.tenantService.AddTenant(context.Background(), &tenant); err != nil {
		// Send a 500 bad_request response
		EncoderHandler(&w, 500, err.Error())
		return
	} else {
		// Send a 201 created response
		EncoderHandler(&w, 201, result)
		logger.Info("Tenants created")
		return
	}
}

func (h tenantHandler) FetchTenantById(w http.ResponseWriter, r *http.Request) {
	var tenant *dto.Tenant
	var err error

	vars := mux.Vars(r)
	TenantID := vars["tenant_id"]
	ID, _ := strconv.ParseInt(TenantID, 10, 32)

	if tenant, err = h.tenantService.FetchTenantById(context.Background(), int32(ID)); err != nil {
		// Send a 500 internal server error response
		EncoderHandler(&w, 500, err.Error())
		return
	} else {
		EncoderHandler(&w, 200, tenant)
		logger.Info("Requested Tenant fetched")
		return
	}
}

func (h tenantHandler) FetchTenantByName(w http.ResponseWriter, r *http.Request) {
	var tenant *dto.Tenant
	var err error

	vars := mux.Vars(r)
	TenantName := vars["tenant_name"]

	if tenant, err = h.tenantService.FetchTenantByName(context.Background(), TenantName); err != nil {
		// Send a 500 internal server error response
		EncoderHandler(&w, 500, err.Error())
		return
	} else {
		EncoderHandler(&w, 200, tenant)
		logger.Info("Requested Tenant fetched")
		return
	}
}

func (h tenantHandler) FetchAllTenants(w http.ResponseWriter, r *http.Request) {
	var tenants []*dto.Tenant
	var err error
	if tenants, err = h.tenantService.FetchAllTenants(context.Background()); err != nil {
		// Send a 500 bad_request response
		EncoderHandler(&w, 500, err.Error())
		return
	} else {
		EncoderHandler(&w, 200, tenants)
		logger.Info("Tenants fetched")
		return
	}
}

// CHANGE THIS TO DEACTIVATE TENANT
func (h tenantHandler) DeactivateTenant(w http.ResponseWriter, r *http.Request) {
	// Read dynamic id parameter
	vars := mux.Vars(r)
	TenantID := vars["tenant_id"]
	ID, _ := strconv.ParseInt(TenantID, 10, 32)

	// user := CookieHandler(r)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 400, err.Error())
		return
	}
	var user model.User
	json.Unmarshal(body, &user)

	if result, err := h.tenantService.DeactivateTenant(context.Background(), int32(ID), user.Email); err != nil {
		// Send a 400 bad_request response
		EncoderHandler(&w, 500, err.Error())
		return
	} else {
		logger.Info("Tenants Deactivated")
		EncoderHandler(&w, 200, result)
		return
	}

}
