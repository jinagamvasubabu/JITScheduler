package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/model"
	"github.com/jinagamvasubabu/JITScheduler-svc/model/dto"
	"github.com/jinagamvasubabu/JITScheduler-svc/repository"
	"go.uber.org/zap"
)

type tenantService struct {
	tenantRepository repository.TenantRepository
}

type TenantService interface {
	AddTenant(ctx context.Context, tenant *dto.Tenant) (string, error)
	DeactivateTenant(ctx context.Context, TenantID int32, user string) (string, error)
	FetchTenantById(ctx context.Context, TenantID int32) (*dto.Tenant, error)
	FetchTenantByName(ctx context.Context, TenantName string) (*dto.Tenant, error)
	FetchAllTenants(ctx context.Context) ([]*dto.Tenant, error)
}

func NewTenantService(ctx context.Context, tenantRepository repository.TenantRepository) TenantService {
	return tenantService{
		tenantRepository: tenantRepository,
	}
}

func (t tenantService) AddTenant(ctx context.Context, tenant *dto.Tenant) (string, error) {
	//create the entity obhect
	properties := dto.Properties{
		Brokers: tenant.Properties.Brokers,
		Topic:   tenant.Properties.Topic,
	}
	props, err := json.Marshal(properties)
	if err != nil {
		return "", err
	}
	//convert dto to entity model
	tenantEntity := &model.Tenant{
		ID:         tenant.ID,
		Name:       tenant.Name,
		Mode:       tenant.Mode,
		Properties: string(props),
		Status:     model.TenantStatus.ENABLED,
		UpdatedBy:  tenant.UpdatedBy,
		UpdatedAt:  time.Now(),
	}
	if err := t.tenantRepository.AddTenant(ctx, tenantEntity); err != nil {
		logger.Error("Error while creating the tenant", zap.Error(err))
		return "", err
	}
	return "success creating the tenant", nil
}

func (t tenantService) FetchTenantById(ctx context.Context, TenantID int32) (*dto.Tenant, error) {
	tenant, err := t.tenantRepository.FetchTenantById(ctx, TenantID)
	if err != nil {
		logger.Error("Error while fetching all the tenants", zap.Error(err))
		return nil, err
	}
	var properties dto.Properties
	json.Unmarshal([]byte(tenant.Properties), &properties)
	tenantDto := &dto.Tenant{
		ID:         tenant.ID,
		Name:       tenant.Name,
		Mode:       tenant.Mode,
		Properties: properties,
		Status:     tenant.Status,
		UpdatedBy:  tenant.UpdatedBy,
		UpdatedAt:  tenant.UpdatedAt,
	}
	return tenantDto, nil
}

func (t tenantService) FetchTenantByName(ctx context.Context, TenantName string) (*dto.Tenant, error) {
	tenant, err := t.tenantRepository.FetchTenantByName(ctx, TenantName)
	if err != nil {
		logger.Error("Error while fetching all the tenants", zap.Error(err))
		return nil, err
	}
	var properties dto.Properties
	json.Unmarshal([]byte(tenant.Properties), &properties)
	tenantDto := &dto.Tenant{
		ID:         tenant.ID,
		Name:       tenant.Name,
		Mode:       tenant.Mode,
		Properties: properties,
		Status:     tenant.Status,
		UpdatedBy:  tenant.UpdatedBy,
		UpdatedAt:  tenant.UpdatedAt,
	}
	return tenantDto, nil
}

func (t tenantService) FetchAllTenants(ctx context.Context) ([]*dto.Tenant, error) {
	var tenants []*model.Tenant
	var err error

	tenants, err = t.tenantRepository.FetchAllTenants(ctx)
	var tenantsDto []*dto.Tenant

	if err != nil {
		logger.Error("Error while fetching all the tenants", zap.Error(err))
		return tenantsDto, err
	}

	for _, tenant := range tenants {
		var properties dto.Properties
		json.Unmarshal([]byte(tenant.Properties), &properties)
		tenantsDto = append(tenantsDto, &dto.Tenant{
			ID:         tenant.ID,
			Name:       tenant.Name,
			Mode:       tenant.Mode,
			Properties: properties,
			Status:     tenant.Status,
			UpdatedBy:  tenant.UpdatedBy,
			UpdatedAt:  tenant.UpdatedAt,
		})
	}
	return tenantsDto, err
}

func (t tenantService) DeactivateTenant(ctx context.Context, TenantID int32, user string) (string, error) {
	//fetch the tenant to send its name
	if err := t.tenantRepository.DeactivateTenant(ctx, TenantID, user); err != nil {
		logger.Error("Error while deactivating the tenant", zap.Error(err))
		return "", err
	}
	return "success deactivating the tenant and its events", nil
}
