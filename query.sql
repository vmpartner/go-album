-- Find duplicates
SELECT count(1) AS cnt, path FROM files GROUP BY hash HAVING cnt > 1;

-- Mimes
SELECT *
FROM files
WHERE mime_type NOT IN
      (
       'image/jpeg',
       'image/png',
       'image/bmp',
       'image/tiff',
       'image/gif',
       'image/vnd.adobe.photoshop',
          --
       'video/x-msvideo',
       'video/mp4',
       'video/x-ms-asf',
       'video/mpeg',
       'video/3gpp',
       'video/quicktime',
          --
       'audio/mpeg',
       'audio/wav',
          --
       'font/ttf',
          --
       'application/pdf',
       'application/zip',
       'application/vnd.ms-excel',
       'application/octet-stream',
       'application/x-rar-compressed',
       'application/rss+xml',
       'application/msword',
          --
       'text/plain; charset=utf-8',
       'text/html; charset=utf-8'
          );