-- Free-form tags on books (1-32 chars of a-z, 0-9 and "-", validated by the API)
CREATE TABLE book_tags (
    book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (book_id, tag)
);

CREATE INDEX idx_book_tags_tag ON book_tags(tag);
