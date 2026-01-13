package business

import (
	"context"
	"fmt"

	"goodreads-clone.com/internal/db"
)

// DefaultShelfNames are the four default shelves created for each new user.
var DefaultShelfNames = []string{
	"To Read",
	"Currently Reading",
	"Finished",
	"DNF",
}

// CreateShelf creates a new custom shelf for a user.
func (b *Business) CreateShelf(ctx context.Context, userID int64, name string) (*db.Shelf, error) {
	shelf, err := b.Storage.Shelves.Create(ctx, userID, name, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create shelf: %w", err)
	}
	return shelf, nil
}

// CreateDefaultShelves creates the 4 default shelves for a user.
func (b *Business) CreateDefaultShelves(ctx context.Context, userID int64) ([]db.Shelf, error) {
	shelves := make([]db.Shelf, 0, len(DefaultShelfNames))
	for _, name := range DefaultShelfNames {
		shelf, err := b.Storage.Shelves.Create(ctx, userID, name, true)
		if err != nil {
			return nil, fmt.Errorf("failed to create default shelf %q: %w", name, err)
		}
		shelves = append(shelves, *shelf)
	}
	return shelves, nil
}

// GetShelfByID retrieves a shelf by its ID.
func (b *Business) GetShelfByID(ctx context.Context, id int64) (*db.Shelf, error) {
	shelf, err := b.Storage.Shelves.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get shelf by ID: %w", err)
	}
	return shelf, nil
}

// GetShelfByUserAndName retrieves a shelf by user ID and name.
func (b *Business) GetShelfByUserAndName(ctx context.Context, userID int64, name string) (*db.Shelf, error) {
	shelf, err := b.Storage.Shelves.GetByUserAndName(ctx, userID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get shelf by user and name: %w", err)
	}
	return shelf, nil
}

// ListShelvesByUser retrieves all shelves for a user.
func (b *Business) ListShelvesByUser(ctx context.Context, userID int64) ([]db.Shelf, error) {
	shelves, err := b.Storage.Shelves.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shelves by user: %w", err)
	}
	return shelves, nil
}

// UpdateShelf updates a custom shelf's name. Cannot update default shelves.
func (b *Business) UpdateShelf(ctx context.Context, id int64, name string) (*db.Shelf, error) {
	shelf, err := b.Storage.Shelves.Update(ctx, id, name)
	if err != nil {
		return nil, fmt.Errorf("failed to update shelf: %w", err)
	}
	return shelf, nil
}

// DeleteShelf deletes a custom shelf. Cannot delete default shelves.
func (b *Business) DeleteShelf(ctx context.Context, id int64) error {
	if err := b.Storage.Shelves.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete shelf: %w", err)
	}
	return nil
}

// GetDefaultShelvesByUser retrieves only the default shelves for a user.
func (b *Business) GetDefaultShelvesByUser(ctx context.Context, userID int64) ([]db.Shelf, error) {
	shelves, err := b.Storage.Shelves.GetDefaultsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default shelves by user: %w", err)
	}
	return shelves, nil
}
