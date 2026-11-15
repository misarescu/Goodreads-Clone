package storage

import (
	"context"
	"fmt"

	"goodreads-clone.com/internal/db"
)

type recommendationRepository struct {
	queries *db.Queries
}

func (r *recommendationRepository) ForUser(
	ctx context.Context, userID, limit int64,
) ([]db.GetRecommendationsRow, error) {
	rows, err := r.queries.GetRecommendations(ctx, db.GetRecommendationsParams{TargetUser: userID, MaxResults: limit})
	if err != nil {
		return nil, fmt.Errorf("get recommendations query failed: %w", err)
	}
	return rows, nil
}
