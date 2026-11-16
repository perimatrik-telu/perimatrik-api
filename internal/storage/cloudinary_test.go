package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"perimatrik-api-go/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestCloudinaryUploadRedirectDelete(t *testing.T) {
	media, err := NewCloudinary(config.CloudinaryConfig{CloudName: "test-cloud", APIKey: "test-key", APISecret: "test-secret", Folder: "perimatrik"})
	if err != nil {
		t.Fatal(err)
	}
	assets := map[string][]byte{}
	var uploads, destroys int
	media.client.Upload.Client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Scheme != "https" || req.URL.Host != "api.cloudinary.com" {
			t.Fatalf("unexpected endpoint: %s", req.URL)
		}
		var result map[string]any
		if strings.HasSuffix(req.URL.Path, "/upload") {
			if err := req.ParseMultipartForm(1 << 20); err != nil {
				t.Fatal(err)
			}
			defer req.MultipartForm.RemoveAll()
			if req.FormValue("overwrite") != "false" || req.FormValue("signature") == "" || req.FormValue("api_key") != "test-key" {
				t.Fatal("upload must be signed and preserve existing assets")
			}
			if req.FormValue("api_secret") != "" {
				t.Fatal("secret sent as form value")
			}
			id := req.FormValue("public_id")
			file, _, err := req.FormFile("file")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err)
			}
			if _, exists := assets[id]; !exists {
				assets[id] = data
			}
			uploads++
			result = map[string]any{"public_id": id, "resource_type": "image", "secure_url": "https://res.cloudinary.com/test-cloud/image/upload/" + id}
		} else if strings.HasSuffix(req.URL.Path, "/destroy") {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			params, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}
			if params.Get("invalidate") != "true" || params.Get("signature") == "" {
				t.Fatal("delete must invalidate CDN and be signed")
			}
			delete(assets, params.Get("public_id"))
			destroys++
			result = map[string]any{"result": "ok"}
		} else {
			t.Fatalf("unexpected Cloudinary call: %s", req.URL)
		}
		body, _ := json.Marshal(result)
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(body)), Request: req}, nil
	})
	root := filepath.Join(t.TempDir(), "must-not-exist")
	store := &Storage{Root: root, AppURL: "https://api.test", cloudinary: media}
	var payload bytes.Buffer
	w := multipart.NewWriter(&payload)
	part, _ := w.CreateFormFile("image", "photo.png")
	_, _ = part.Write([]byte("\x89PNG\r\n\x1a\nbackup"))
	_ = w.Close()
	form, err := multipart.NewReader(&payload, w.Boundary()).ReadForm(1024)
	if err != nil {
		t.Fatal(err)
	}
	defer form.RemoveAll()
	imageURL, err := store.SaveURL(form.File["image"][0], ArticleDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(imageURL, "https://api.test/storage/articles/") {
		t.Fatalf("unstable public URL: %s", imageURL)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("Cloudinary upload wrote to local disk")
	}
	rel := strings.TrimPrefix(imageURL, "https://api.test/storage/")
	id, _ := media.publicID(rel)
	if len(assets[id]) == 0 {
		t.Fatal("asset not uploaded")
	}
	for _, method := range []string{"GET", "HEAD"} {
		response := httptest.NewRecorder()
		store.ServeHTTP(response, httptest.NewRequest(method, "/storage/"+rel, nil), rel)
		if response.Code != 302 || !strings.HasPrefix(response.Header().Get("Location"), "https://res.cloudinary.com/test-cloud/image/upload/") || !strings.Contains(response.Header().Get("Location"), id) {
			t.Fatalf("CDN redirect: %d %v", response.Code, response.Header())
		}
		if strings.Contains(response.Header().Get("Location"), "test-secret") {
			t.Fatal("secret in public URL")
		}
	}
	store.DeleteByURL(ArticleDir, imageURL)
	if _, exists := assets[id]; exists || destroys != 1 {
		t.Fatal("asset not deleted")
	}
	// Deterministic IDs preserve the original legacy filenames and do not collide.
	ids := map[string]bool{}
	for _, key := range []string{"articles/legacy.", "articles/legacy.jpg", "articles/legacy.png", "galleries/legacy.jpg"} {
		id, err := media.publicID(key)
		if err != nil {
			t.Fatal(err)
		}
		if ids[id] {
			t.Fatal("filename collision")
		}
		ids[id] = true
		if err := media.Put(context.Background(), key, strings.NewReader("original")); err != nil {
			t.Fatal(err)
		}
		if err := media.Put(context.Background(), key, strings.NewReader("replacement")); err != nil {
			t.Fatal(err)
		}
		if string(assets[id]) != "original" {
			t.Fatal("existing asset overwritten")
		}
	}
	if uploads != 9 {
		t.Fatalf("unexpected upload count: %d", uploads)
	}
	for _, key := range []string{"../.env", "articles/../secret.png", "articles/a/b.png", "articles/.env", "other/photo.png", "articles/a\\b.png"} {
		response := httptest.NewRecorder()
		store.ServeHTTP(response, httptest.NewRequest("GET", "/", nil), key)
		if response.Code != 404 {
			t.Fatalf("unsafe key accepted: %s", key)
		}
	}
}

func TestCloudinaryConfiguration(t *testing.T) {
	cfg := config.CloudinaryConfig{CloudName: "demo", APIKey: "key", APISecret: "secret"}
	media, err := NewCloudinary(cfg)
	if err != nil || media.folder != "perimatrik" {
		t.Fatalf("default folder: %v", err)
	}
	preview := cfg
	preview.Folder = "perimatrik-preview"
	other, err := NewCloudinary(preview)
	if err != nil {
		t.Fatal(err)
	}
	mainID, _ := media.publicID("articles/a.png")
	previewID, _ := other.publicID("articles/a.png")
	if mainID == previewID {
		t.Fatal("preview and production share the same asset ID")
	}
	for _, invalid := range []config.CloudinaryConfig{
		{CloudName: "demo", APIKey: "key"},
		{CloudName: "demo", APIKey: "key", APISecret: "secret", Folder: "../bad"},
		{CloudName: "https://host", APIKey: "key", APISecret: "secret"},
	} {
		if _, err := NewCloudinary(invalid); err == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
	if _, err := New(&config.Config{StorageDriver: "r2"}); err == nil {
		t.Fatal("obsolete driver silently accepted")
	}
}
