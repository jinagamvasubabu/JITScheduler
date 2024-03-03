package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Buddy-Git/JITScheduler-svc/model"

	"gorm.io/gorm"
)

type tenantRepository struct {
	db *gorm.DB
}

type TenantRepository interface {
	AddTenant(ctx context.Context, tenant *model.Tenant) error
	DeactivateTenant(ctx context.Context, TenantID int32, user string) error
	FetchTenantById(ctx context.Context, TenantID int32) (*model.Tenant, error)
	FetchTenantByName(ctx context.Context, TenantName string) (*model.Tenant, error)
	FetchAllTenants(ctx context.Context) ([]*model.Tenant, error)
}

func NewTenantRepository(ctx context.Context, db *gorm.DB) *tenantRepository {
	return &tenantRepository{
		db: db,
	}
}

func (t *tenantRepository) AddTenant(ctx context.Context, tenant *model.Tenant) error {
	existingTenant := &model.Tenant{}
	if err := t.db.Where("name = ?", tenant.Name).First(existingTenant).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existingTenant.ID != 0 {
		if existingTenant.Status == "ENABLED" {
			return errors.New("tenant has already registered")
		}
		tenant.ID = existingTenant.ID
		return t.db.Where("name = ?", tenant.Name).Save(tenant).Error
	}
	return t.db.Create(tenant).Error
}

func (t *tenantRepository) DeactivateTenant(ctx context.Context, TenantID int32, user string) error {
	existingTenant := &model.Tenant{}
	if db := t.db.Where("id", TenantID).First(existingTenant).
		Updates(map[string]interface{}{"status": "DISABLED", "updated_by": user, "updated_at": time.Now()}); db.Error != nil {
		return db.Error
	}
	// CANCEL ALL EVENTS
	if db := t.db.Model(&model.Event{TenantID: TenantID}).
		Updates(model.Event{Status: "CANCELLED", UpdatedBy: user, UpdatedAt: time.Now()}); db.Error != nil {
		return db.Error
	}
	return nil
}

func (t *tenantRepository) FetchTenantById(ctx context.Context, TenantID int32) (*model.Tenant, error) {
	tenant := &model.Tenant{}
	if err := t.db.Where("id = ?", TenantID).First(tenant).Error; err != nil {
		return nil, err
	}
	return tenant, nil
}

func (t *tenantRepository) FetchTenantByName(ctx context.Context, TenantName string) (*model.Tenant, error) {
	tenant := &model.Tenant{}
	if err := t.db.Where("name = ?", TenantName).First(tenant).Error; err != nil {
		return nil, err
	}
	return tenant, nil
}

func (t *tenantRepository) FetchAllTenants(ctx context.Context) ([]*model.Tenant, error) {
	var tenants []*model.Tenant
	if err := t.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	// fmt.Println(tenants)
	return tenants, nil
}
