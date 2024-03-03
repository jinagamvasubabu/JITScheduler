package handler

import (
	"github.com/gorilla/mux"
)

func InitRouter(t tenantHandler, e eventHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/tenants", t.FetchAllTenants).Methods("GET", "OPTIONS")
	router.HandleFunc("/tenants/{tenant_id}", t.FetchTenantById).Methods("GET", "OPTIONS")
	router.HandleFunc("/tenants/{tenant_name}", t.FetchTenantByName).Methods("GET", "OPTIONS")
	router.HandleFunc("/tenants/{tenant_id}", t.DeactivateTenant).Methods("PUT", "OPTIONS")
	router.HandleFunc("/tenants", t.AddTenant).Methods("POST", "OPTIONS")
	router.HandleFunc("/tenants/{tenant_id}/events", e.FetchAllEventsUnderTenant).Methods("GET", "OPTIONS")
	router.HandleFunc("/tenants/{tenant_id}/events/{event_id}", e.DeactivateEvent).Methods("PUT", "OPTIONS")
	router.HandleFunc("/tenants/{tenant_id}/events", e.AddEvent).Methods("POST", "OPTIONS")
	return router
}
