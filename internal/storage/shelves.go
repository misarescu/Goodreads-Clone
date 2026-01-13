package storage

import (
	"context"
	"fmt"

	"goodreads-clone.com/internal/db"
)

type shelfRepository struct {
	queries *db.Queries
}

func (r *shelfRepository) Create(ctx context.Context, userID int64, name string, isDefault bool) (*db.Shelf, error) {
	shelf, err := r.queries.CreateShelf(ctx, db.CreateShelfParams{
		UserID:    userID,
		Name:      name,
		IsDefault: isDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("create shelf query failed: %w", err)
	}
	return &shelf, nil
}

func (r *shelfRepository) GetByID(ctx context.Context, id int64) (*db.Shelf, error) {
	shelf, err := r.queries.GetShelfByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get shelf by ID query failed: %w", err)
	}
	return &shelf, nil
}

func (r *shelfRepository) GetByUserAndName(ctx context.Context, userID int64, name string) (*db.Shelf, error) {
	shelf, err := r.queries.GetShelfByUserAndName(ctx, db.GetShelfByUserAndNameParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		return nil, fmt.Errorf("get shelf by user and name query failed: %w", err)
	}
	return &shelf, nil
}

func (r *shelfRepository) ListByUser(ctx context.Context, userID int64) ([]db.Shelf, error) {
	shelves, err := r.queries.ListShelvesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list shelves by user query failed: %w", err)
	}
	return shelves, nil
}

func (r *shelfRepository) Update(ctx context.Context, id int64, name string) (*db.Shelf, error) {
	shelf, err := r.queries.UpdateShelf(ctx, db.UpdateShelfParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("update shelf query failed: %w", err)
	}
	return &shelf, nil
}

func (r *shelfRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteShelf(ctx, id); err != nil {
		return fmt.Errorf("delete shelf query failed: %w", err)
	}
	return nil
}

func (r *shelfRepository) GetDefaultsByUser(ctx context.Context, userID int64) ([]db.Shelf, error) {
	shelves, err := r.queries.GetDefaultShelvesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get default shelves by user query failed: %w", err)
	}
	return shelves, nil
}
