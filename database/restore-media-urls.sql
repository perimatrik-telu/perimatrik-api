-- Run in the application database only after uploading backup files to Cloudinary.
-- Preview the affected counts first. Only the old storage URL prefix is replaced.
SET @old_media = 'https://api.perimatrik.org/storage/';
SET @new_media = 'https://perimatrik-api.vercel.app/storage/';

SELECT 'articles' AS source, COUNT(*) AS affected FROM articles WHERE image LIKE CONCAT(@old_media, '%')
UNION ALL SELECT 'news', COUNT(*) FROM news WHERE image LIKE CONCAT(@old_media, '%')
UNION ALL SELECT 'galleries', COUNT(*) FROM galleries WHERE image LIKE CONCAT(@old_media, '%')
UNION ALL SELECT 'users', COUNT(*) FROM users WHERE avatar LIKE CONCAT(@old_media, '%')
UNION ALL SELECT 'members', COUNT(*) FROM members WHERE avatar LIKE CONCAT(@old_media, '%');

-- Execute this block separately after checking the counts above.
START TRANSACTION;
UPDATE articles SET image = REPLACE(image, @old_media, @new_media) WHERE image LIKE CONCAT(@old_media, '%');
UPDATE news SET image = REPLACE(image, @old_media, @new_media) WHERE image LIKE CONCAT(@old_media, '%');
UPDATE galleries SET image = REPLACE(image, @old_media, @new_media) WHERE image LIKE CONCAT(@old_media, '%');
UPDATE users SET avatar = REPLACE(avatar, @old_media, @new_media) WHERE avatar LIKE CONCAT(@old_media, '%');
UPDATE members SET avatar = REPLACE(avatar, @old_media, @new_media) WHERE avatar LIKE CONCAT(@old_media, '%');
COMMIT;
