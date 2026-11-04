-- +goose Up
-- Posts written before this migration have no FK, so posts whose feed was
-- deleted (e.g. by `gator reset`) are orphaned. Remove them so the FK can be added.
DELETE FROM posts p
WHERE NOT EXISTS (SELECT 1 FROM feeds f WHERE f.id = p.feed_id);

ALTER TABLE posts
ADD CONSTRAINT fk_posts_feed FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE;

-- Serves GetPostsForUser: posts per followed feed, newest first.
CREATE INDEX posts_feed_published_idx ON posts (feed_id, published_at DESC);

-- +goose Down
DROP INDEX posts_feed_published_idx;
ALTER TABLE posts DROP CONSTRAINT fk_posts_feed;
