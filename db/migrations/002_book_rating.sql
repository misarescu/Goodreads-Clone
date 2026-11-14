-- Book rating: 0 to 5 in increments of 0.25, NULL when unrated
ALTER TABLE books ADD COLUMN rating REAL CHECK (rating IS NULL OR (rating >= 0 AND rating <= 5 AND rating * 4 = CAST(rating * 4 AS INTEGER)));
