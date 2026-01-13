package business

import (
	"context"
	"fmt"

	"goodreads-clone.com/internal/db"
)

// CreateUser creates a new user. Password should already be hashed.
func (b *Business) CreateUser(ctx context.Context, email, username, passwordHash string) (*db.User, error) {
	user, err := b.Storage.Users.Create(ctx, email, username, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID.
func (b *Business) GetUserByID(ctx context.Context, id int64) (*db.User, error) {
	user, err := b.Storage.Users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by their email.
func (b *Business) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := b.Storage.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// GetUserByUsername retrieves a user by their username.
func (b *Business) GetUserByUsername(ctx context.Context, username string) (*db.User, error) {
	user, err := b.Storage.Users.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// UpdateUser updates a user's email and username.
func (b *Business) UpdateUser(ctx context.Context, id int64, email, username string) (*db.User, error) {
	user, err := b.Storage.Users.Update(ctx, id, email, username)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// DeleteUser deletes a user by their ID.
func (b *Business) DeleteUser(ctx context.Context, id int64) error {
	if err := b.Storage.Users.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers retrieves a paginated list of users.
func (b *Business) ListUsers(ctx context.Context, limit, offset int64) ([]db.User, error) {
	users, err := b.Storage.Users.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}
