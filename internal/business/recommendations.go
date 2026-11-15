package business

import (
	"context"
	"encoding/json"
	"fmt"
)

// Recommendation limits.
const (
	DefaultRecommendationLimit = 10
	MaxRecommendationLimit     = 50
)

// Recommendation is a suggested book with why it was suggested.
type Recommendation struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Author  string   `json:"author"`
	Score   int64    `json:"score"`
	Because []string `json:"because"`
}

// GetRecommendations returns up to limit books finished by readers similar to
// the user. It returns an error wrapping sql.ErrNoRows when the user does not
// exist; a user who finished nothing gets an empty list.
func (b *Business) GetRecommendations(ctx context.Context, userID, limit int64) ([]Recommendation, error) {
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to get user for recommendations: %w", err)
	}
	rows, err := b.Storage.Recommendations.ForUser(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}
	out := make([]Recommendation, 0, len(rows))
	for _, r := range rows {
		because := []string{}
		if err := json.Unmarshal([]byte(r.BecauseTitles), &because); err != nil {
			return nil, fmt.Errorf("failed to decode because titles: %w", err)
		}
		out = append(out, Recommendation{ID: r.ID, Title: r.Title, Author: r.Author, Score: r.Score, Because: because})
	}
	return out, nil
}
