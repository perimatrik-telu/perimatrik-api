// Package storage is the "public" disk: uploaded files live under Root and are
// publicly reachable at {AppURL}/storage/{path}.
package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"perimatrik-api-go/internal/config"
	"perimatrik-api-go/internal/support"
	"perimatrik-api-go/internal/validation"
)

// Upload directories (config/path.php).
const (
	ArticleDir = "articles"
	NewsDir    = "news"
	GalleryDir = "galleries"
	MemberDir  = "member-avatars"
	AvatarDir  = "avatars"
)

var safeExt = map[string]bool{"jpg": true, "jpeg": true, "png": true, "gif": true, "svg": true, "webp": true}

type Storage struct {
	Root       string
	AppURL     string
	cloudinary *Cloudinary
}

// New selects Cloudinary storage or the local public disk.
func New(cfg *config.Config) (*Storage, error) {
	s := &Storage{Root: cfg.StoragePath, AppURL: cfg.AppURL}
	switch cfg.StorageDriver {
	case "", "local":
	case "cloudinary":
		var err error
		s.cloudinary, err = NewCloudinary(cfg.Cloudinary)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid STORAGE_DRIVER: expected local or cloudinary")
	}
	return s, nil
}

// Save stores an upload as {dir}/{random30}.{ext} and returns the filename.
func (s *Storage) Save(fh *multipart.FileHeader, dir string) (string, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fh.Filename), "."))
	if !safeExt[ext] {
		// Never trust a client extension that could be served as HTML/JS.
		ext = validation.DetectExtension(fh)
	}
	name := support.RandomString(30) + "." + ext

	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	if s.cloudinary != nil {
		ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
		defer cancel()
		if err := s.cloudinary.Put(ctx, dir+"/"+name, src); err != nil {
			return "", err
		}
		return name, nil
	}
	if err := os.MkdirAll(filepath.Join(s.Root, dir), 0o755); err != nil {
		return "", err
	}

	dst, err := os.Create(filepath.Join(s.Root, dir, name))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return "", err
	}
	return name, dst.Close()
}

// URL is the public URL of {dir}/{filename}.
func (s *Storage) URL(dir, filename string) string {
	return s.AppURL + "/storage/" + strings.Trim(dir, "/") + "/" + filename
}

// SaveURL saves the upload and returns its public URL.
func (s *Storage) SaveURL(fh *multipart.FileHeader, dir string) (string, error) {
	name, err := s.Save(fh, dir)
	if err != nil {
		return "", err
	}
	return s.URL(dir, name), nil
}

// DeleteByURL deletes {dir}/{basename of the URL path}; missing files are ignored.
func (s *Storage) DeleteByURL(dir, fileURL string) {
	p := fileURL
	if u, err := url.Parse(fileURL); err == nil {
		p = u.Path
	}
	base := path.Base(p)
	if base == "." || base == "/" || base == "" {
		return
	}
	if s.cloudinary != nil {
		ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
		defer cancel()
		if err := s.cloudinary.Delete(ctx, dir+"/"+base); err != nil {
			log.Printf("storage delete: %v", err)
		}
		return
	}
	_ = os.Remove(filepath.Join(s.Root, dir, base))
}

// Resolve maps a relative public-disk path to a file path, refusing anything
// that escapes Root. ok is false if it does not exist or is a directory.
func (s *Storage) Resolve(rel string) (string, bool) {
	clean := path.Clean("/" + rel)
	full := filepath.Join(s.Root, filepath.FromSlash(clean))
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return "", false
	}
	abs, err := filepath.Abs(full)
	if err != nil || !strings.HasPrefix(abs, root+string(filepath.Separator)) {
		return "", false
	}
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		return "", false
	}
	return abs, true
}
