package business

import (
	"context"

	"goodreads-clone.com/internal/db"
)

// Default shelf names
var DefaultShelfNames = []string{
	"To Read",
	"Currently Reading",
	"Finished",
	"DNF",
}

// CreateShelf creates a new custom shelf for a user.
func (b *Business) CreateShelf(ctx context.Context, userID int64, name string) (*db.Shelf, error) {
	return b.Storage.Shelves.Create(ctx, userID, name, false)
}

// CreateDefaultShelves creates the 4 default shelves for a user.
func (b *Business) CreateDefaultShelves(ctx context.Context, userID int64) ([]db.Shelf, error) {
	shelves := make([]db.Shelf, 0, len(DefaultShelfNames))
	for _, name := range DefaultShelfNames {
		shelf, err := b.Storage.Shelves.Create(ctx, userID, name, true)
		if err != nil {
			return nil, err
		}
		shelves = append(shelves, *shelf)
	}
	return shelves, nil
}

// GetShelfByID retrieves a shelf by its ID.
func (b *Business) GetShelfByID(ctx context.Context, id int64) (*db.Shelf, error) {
	return b.Storage.Shelves.GetByID(ctx, id)
}

// GetShelfByUserAndName retrieves a shelf by user ID and name.
func (b *Business) GetShelfByUserAndName(ctx context.Context, userID int64, name string) (*db.Shelf, error) {
	return b.Storage.Shelves.GetByUserAndName(ctx, userID, name)
}

// ListShelvesByUser retrieves all shelves for a user.
func (b *Business) ListShelvesByUser(ctx context.Context, userID int64) ([]db.Shelf, error) {
	return b.Storage.Shelves.ListByUser(ctx, userID)
}

// UpdateShelf updates a custom shelf's name. Cannot update default shelves.
func (b *Business) UpdateShelf(ctx context.Context, id int64, name string) (*db.Shelf, error) {
	return b.Storage.Shelves.Update(ctx, id, name)
}

// DeleteShelf deletes a custom shelf. Cannot delete default shelves.
func (b *Business) DeleteShelf(ctx context.Context, id int64) error {
	return b.Storage.Shelves.Delete(ctx, id)
}

// GetDefaultShelvesByUser retrieves only the default shelves for a user.
func (b *Business) GetDefaultShelvesByUser(ctx context.Context, userID int64) ([]db.Shelf, error) {
	return b.Storage.Shelves.GetDefaultsByUser(ctx, userID)
}
