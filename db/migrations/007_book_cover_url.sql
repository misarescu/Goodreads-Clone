-- Optional cover image URL (absolute http(s) URL, validated by the API), NULL when unset
ALTER TABLE books ADD COLUMN cover_url TEXT;
