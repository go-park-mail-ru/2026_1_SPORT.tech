DELETE FROM content_post_like
WHERE (post_id = 2001 AND user_id = 1002)
   OR (post_id = 2003 AND user_id = 1001);

DELETE FROM content_post_block
WHERE post_block_id IN (2101, 2102, 2103, 2104, 2105, 2106);

DELETE FROM content_post
WHERE post_id IN (2001, 2002, 2003);
