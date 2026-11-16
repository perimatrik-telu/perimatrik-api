package storage

import (
	"path"
	"strings"
)

// MediaKey only allows files in the application's public media directories.
func MediaKey(rel string) (string, bool) {
	key := strings.TrimPrefix(rel, "/")
	parts := strings.Split(key, "/")
	if len(parts) != 2 || parts[1] == "" || strings.ContainsAny(key, "\\\x00") {
		return "", false
	}
	switch parts[0] {
	case ArticleDir, NewsDir, GalleryDir, MemberDir, AvatarDir:
	default:
		return "", false
	}
	ext := path.Ext(parts[1])
	if parts[1] == "." || parts[1] == ".." || strings.HasPrefix(parts[1], ".") {
		return "", false
	}
	if ext != "" && ext != "." && !safeExt[strings.ToLower(strings.TrimPrefix(ext, "."))] {
		return "", false
	}
	return key, true
}
