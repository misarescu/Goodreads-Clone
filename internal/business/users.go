package business

import (
	"context"

	"goodreads-clone.com/internal/db"
)

// CreateUser creates a new user. Password should already be hashed.
func (b *Business) CreateUser(ctx context.Context, email, username, passwordHash string) (*db.User, error) {
	return b.Storage.Users.Create(ctx, email, username, passwordHash)
}

// GetUserByID retrieves a user by their ID.
func (b *Business) GetUserByID(ctx context.Context, id int64) (*db.User, error) {
	return b.Storage.Users.GetByID(ctx, id)
}

// GetUserByEmail retrieves a user by their email.
func (b *Business) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	return b.Storage.Users.GetByEmail(ctx, email)
}

// GetUserByUsername retrieves a user by their username.
func (b *Business) GetUserByUsername(ctx context.Context, username string) (*db.User, error) {
	return b.Storage.Users.GetByUsername(ctx, username)
}

// UpdateUser updates a user's email and username.
func (b *Business) UpdateUser(ctx context.Context, id int64, email, username string) (*db.User, error) {
	return b.Storage.Users.Update(ctx, id, email, username)
}

// DeleteUser deletes a user by their ID.
func (b *Business) DeleteUser(ctx context.Context, id int64) error {
	return b.Storage.Users.Delete(ctx, id)
}

// ListUsers retrieves a paginated list of users.
func (b *Business) ListUsers(ctx context.Context, limit, offset int64) ([]db.User, error) {
	return b.Storage.Users.List(ctx, limit, offset)
}
