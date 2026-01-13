package storage

import (
	"context"
	"fmt"

	"goodreads-clone.com/internal/db"
)

type userRepository struct {
	queries *db.Queries
}

func (r *userRepository) Create(ctx context.Context, email, username, passwordHash string) (*db.User, error) {
	user, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("create user query failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*db.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by ID query failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email query failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username query failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, id int64, email, username string) (*db.User, error) {
	user, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:       id,
		Email:    email,
		Username: username,
	})
	if err != nil {
		return nil, fmt.Errorf("update user query failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user query failed: %w", err)
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int64) ([]db.User, error) {
	users, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list users query failed: %w", err)
	}
	return users, nil
}
