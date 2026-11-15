-- When the book was finished: set the first time progress reaches 100, NULL otherwise
ALTER TABLE shelf_books ADD COLUMN finished_at DATETIME;
