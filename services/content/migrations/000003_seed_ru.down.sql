DELETE FROM content_post_like
WHERE (post_id BETWEEN 2004 AND 2011)
   OR (post_id = 2001 AND user_id = 1008)
   OR (post_id = 2003 AND user_id = 1008);

DELETE FROM content_post_block
WHERE post_block_id BETWEEN 2107 AND 2133;

DELETE FROM content_post
WHERE post_id BETWEEN 2004 AND 2011;
