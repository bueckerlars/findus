package repository

import (
	"context"
	"time"

	"findus/backend/internal/domain"
)

type UserRepository interface {
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, u *domain.User) error
}

type LocationRepository interface {
	Create(ctx context.Context, l *domain.Location) error
	GetByID(ctx context.Context, id string) (*domain.Location, error)
	GetByQRToken(ctx context.Context, token string) (*domain.Location, error)
	Update(ctx context.Context, l *domain.Location) error
	Delete(ctx context.Context, id string) error
	ListChildren(ctx context.Context, parentID *string) ([]domain.Location, error)
	PathFromRoot(ctx context.Context, id string) ([]domain.LocationPathElement, error)
	ListRecentByUpdated(ctx context.Context, limit int) ([]domain.Location, error)
	ListAll(ctx context.Context, limit int) ([]domain.Location, error)
	// ChildCountsByParentID returns how many locations have each parent_id (direct children only).
	ChildCountsByParentID(ctx context.Context) (map[string]int64, error)
	Count(ctx context.Context) (int64, error)
}

type ItemRepository interface {
	Create(ctx context.Context, it *domain.Item) error
	GetByID(ctx context.Context, id string) (*domain.Item, error)
	GetByQRToken(ctx context.Context, token string) (*domain.Item, error)
	Update(ctx context.Context, it *domain.Item) error
	Delete(ctx context.Context, id string) error
	ListByLocation(ctx context.Context, locationID string) ([]domain.Item, error)
	Search(ctx context.Context, q string, limit int) ([]domain.Item, error)
	ListAll(ctx context.Context, limit int) ([]domain.Item, error)
	ListRecentByUpdated(ctx context.Context, limit int) ([]domain.Item, error)
	Count(ctx context.Context) (int64, error)
	ReplaceItemLabels(ctx context.Context, itemID string, labelIDs []string) error
	ListLabelsForItem(ctx context.Context, itemID string) ([]domain.Label, error)
	CountByTemplateType(ctx context.Context, templateType string) (int64, error)
	ReassignTemplateType(ctx context.Context, fromID, toID string) error
	UpdateItemSearchDenorm(ctx context.Context, itemID, searchLabels, searchLocation, searchBody string) error
	UpdateSearchLocationForItemsAtLocation(ctx context.Context, locationID, searchLocation string) error
	// MigrateItemPrimaryKeys runs inside a new DB transaction (PRAGMA foreign_keys=OFF) and updates item_labels then items.
	MigrateItemPrimaryKeys(ctx context.Context, rows []domain.ItemIDMigration) error
}

type ItemTemplateRepository interface {
	List(ctx context.Context) ([]domain.ItemTemplate, error)
	GetByID(ctx context.Context, id string) (*domain.ItemTemplate, error)
	Create(ctx context.Context, t *domain.ItemTemplate) error
	Update(ctx context.Context, t *domain.ItemTemplate) error
	Delete(ctx context.Context, id string) error
}

type InviteRepository interface {
	Create(ctx context.Context, inv *domain.Invite) error
	GetByToken(ctx context.Context, token string) (*domain.Invite, error)
	MarkUsed(ctx context.Context, id string, at time.Time) error
	ListRecent(ctx context.Context, limit int) ([]domain.Invite, error)
}

type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string) error
}
