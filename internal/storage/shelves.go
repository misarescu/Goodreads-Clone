package storage

import (
	"context"

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
		return nil, err
	}
	return &shelf, nil
}

func (r *shelfRepository) GetByID(ctx context.Context, id int64) (*db.Shelf, error) {
	shelf, err := r.queries.GetShelfByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &shelf, nil
}

func (r *shelfRepository) GetByUserAndName(ctx context.Context, userID int64, name string) (*db.Shelf, error) {
	shelf, err := r.queries.GetShelfByUserAndName(ctx, db.GetShelfByUserAndNameParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		return nil, err
	}
	return &shelf, nil
}

func (r *shelfRepository) ListByUser(ctx context.Context, userID int64) ([]db.Shelf, error) {
	return r.queries.ListShelvesByUser(ctx, userID)
}

func (r *shelfRepository) Update(ctx context.Context, id int64, name string) (*db.Shelf, error) {
	shelf, err := r.queries.UpdateShelf(ctx, db.UpdateShelfParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	return &shelf, nil
}

func (r *shelfRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteShelf(ctx, id)
}

func (r *shelfRepository) GetDefaultsByUser(ctx context.Context, userID int64) ([]db.Shelf, error) {
	return r.queries.GetDefaultShelvesByUser(ctx, userID)
}
