-- Full-text search over books (title and author), kept in sync by triggers.
-- Idempotent: `make migrate` re-applies every file.
CREATE VIRTUAL TABLE IF NOT EXISTS books_fts USING fts5(
    title,
    author,
    content='books',
    content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS books_fts_ai AFTER INSERT ON books BEGIN
    INSERT INTO books_fts(rowid, title, author) VALUES (new.id, new.title, new.author);
END;

CREATE TRIGGER IF NOT EXISTS books_fts_ad AFTER DELETE ON books BEGIN
    INSERT INTO books_fts(books_fts, rowid, title, author) VALUES ('delete', old.id, old.title, old.author);
END;

CREATE TRIGGER IF NOT EXISTS books_fts_au AFTER UPDATE ON books BEGIN
    INSERT INTO books_fts(books_fts, rowid, title, author) VALUES ('delete', old.id, old.title, old.author);
    INSERT INTO books_fts(rowid, title, author) VALUES (new.id, new.title, new.author);
END;

-- Fill from the existing rows.
INSERT INTO books_fts(books_fts) VALUES ('rebuild');
