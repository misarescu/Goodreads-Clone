-- name: GetRecommendations :many
-- Books finished by readers similar to the user (who finished at least one
-- book the user finished), excluding the user's shelved books. score is the
-- number of similar readers who finished the book; ties go to the book with
-- more finishers overall, then to the title. because_titles is a JSON array
-- of up to 3 of the user's finished books that recommending readers also
-- finished.
SELECT
    b.id,
    b.title,
    b.author,
    CAST(COUNT(DISTINCT rs.user_id) AS INTEGER) AS score,
    CAST((
        SELECT json_group_array(t.title) FROM (
            SELECT DISTINCT mb.title AS title
            FROM shelf_books msb
            JOIN shelves ms ON ms.id = msb.shelf_id
            JOIN books mb ON mb.id = msb.book_id
            JOIN shelves os ON os.user_id <> ms.user_id
            JOIN shelf_books osb ON osb.shelf_id = os.id AND osb.book_id = msb.book_id
            JOIN shelves cs ON cs.user_id = os.user_id
            JOIN shelf_books csb ON csb.shelf_id = cs.id
            WHERE ms.user_id = @target_user
              AND msb.finished_at IS NOT NULL
              AND osb.finished_at IS NOT NULL
              AND csb.finished_at IS NOT NULL
              AND csb.book_id = b.id
            ORDER BY mb.title, mb.id
            LIMIT 3
        ) t
    ) AS TEXT) AS because_titles
FROM (
    SELECT DISTINCT rsh.user_id AS user_id, rsb.book_id AS book_id
    FROM shelf_books rsb
    JOIN shelves rsh ON rsh.id = rsb.shelf_id
    WHERE rsb.finished_at IS NOT NULL
      AND rsh.user_id <> @target_user
      AND rsh.user_id IN (
          SELECT us.user_id
          FROM shelves us
          JOIN shelf_books usb ON usb.shelf_id = us.id
          WHERE usb.finished_at IS NOT NULL
            AND usb.book_id IN (
                SELECT fsb.book_id
                FROM shelf_books fsb
                JOIN shelves fs ON fs.id = fsb.shelf_id
                WHERE fs.user_id = @target_user
                  AND fsb.finished_at IS NOT NULL
            )
      )
) rs
JOIN books b ON b.id = rs.book_id
WHERE b.id NOT IN (
    SELECT ssb.book_id
    FROM shelf_books ssb
    JOIN shelves ss ON ss.id = ssb.shelf_id
    WHERE ss.user_id = @target_user
)
GROUP BY b.id, b.title, b.author
ORDER BY score DESC,
         (SELECT COUNT(DISTINCT fo.user_id)
          FROM shelf_books fob
          JOIN shelves fo ON fo.id = fob.shelf_id
          WHERE fob.book_id = b.id AND fob.finished_at IS NOT NULL) DESC,
         b.title ASC,
         b.id ASC
LIMIT @max_results;
