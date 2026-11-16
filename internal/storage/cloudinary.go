package storage

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"perimatrik-api-go/internal/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const operationTimeout = 60 * time.Second

type Cloudinary struct {
	client *cloudinary.Cloudinary
	folder string
}

func NewCloudinary(cfg config.CloudinaryConfig) (*Cloudinary, error) {
	for _, setting := range []struct{ name, value string }{
		{"CLOUDINARY_CLOUD_NAME", cfg.CloudName}, {"CLOUDINARY_API_KEY", cfg.APIKey},
		{"CLOUDINARY_API_SECRET", cfg.APISecret},
	} {
		if strings.TrimSpace(setting.value) == "" {
			return nil, fmt.Errorf("%s is required for STORAGE_DRIVER=cloudinary", setting.name)
		}
	}
	if cfg.Folder == "" {
		cfg.Folder = "perimatrik"
	}
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(cfg.CloudName) || !validName.MatchString(cfg.Folder) {
		return nil, fmt.Errorf("CLOUDINARY_CLOUD_NAME and CLOUDINARY_FOLDER must contain only letters, numbers, hyphens or underscores")
	}
	client, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("invalid Cloudinary configuration")
	}
	client.Config.URL.Secure = true
	client.Upload.Client.Timeout = operationTimeout
	return &Cloudinary{client: client, folder: cfg.Folder}, nil
}

// publicID encodes the filename to avoid Cloudinary's extension rules and keep
// legacy names ending in a dot distinct from names with image extensions.
func (c *Cloudinary) publicID(key string) (string, error) {
	key, ok := MediaKey(key)
	if !ok {
		return "", fmt.Errorf("invalid media object key")
	}
	dir, name, _ := strings.Cut(key, "/")
	return c.folder + "/" + dir + "/" + base64.RawURLEncoding.EncodeToString([]byte(name)), nil
}

// Put is signed on the server. Repeated imports never overwrite an existing ID.
func (c *Cloudinary) Put(ctx context.Context, key string, body io.Reader) error {
	id, err := c.publicID(key)
	if err != nil {
		return err
	}
	result, err := c.client.Upload.Upload(ctx, body, uploader.UploadParams{
		PublicID: id, ResourceType: "image", Type: api.Upload,
		Overwrite: api.Bool(false), UniqueFilename: api.Bool(false),
		AllowedFormats: api.CldAPIArray{"jpg", "jpeg", "png", "gif", "webp", "svg"},
	})
	if err != nil {
		return err
	}
	if result.Error.Message != "" {
		return fmt.Errorf("Cloudinary upload: %s", result.Error.Message)
	}
	if result.PublicID != id {
		return fmt.Errorf("Cloudinary returned an unexpected asset ID")
	}
	return nil
}

func (c *Cloudinary) Delete(ctx context.Context, key string) error {
	id, err := c.publicID(key)
	if err != nil {
		return err
	}
	result, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: id, ResourceType: "image", Type: "upload", Invalidate: api.Bool(true),
	})
	if err != nil {
		return err
	}
	if result.Result != "ok" && result.Result != "not found" {
		return fmt.Errorf("Cloudinary delete failed")
	}
	return nil
}

func (c *Cloudinary) URL(key string) (string, error) {
	id, err := c.publicID(key)
	if err != nil {
		return "", err
	}
	image, err := c.client.Image(id)
	if err != nil {
		return "", err
	}
	return image.String()
}

// ServeHTTP preserves existing media routes and delivers image bytes via CDN.
func (s *Storage) ServeHTTP(w http.ResponseWriter, req *http.Request, rel string) {
	if s.cloudinary == nil {
		p, ok := s.Resolve(rel)
		if !ok {
			http.NotFound(w, req)
			return
		}
		http.ServeFile(w, req, p)
		return
	}
	target, err := s.cloudinary.URL(rel)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.Redirect(w, req, target, http.StatusFound)
}
