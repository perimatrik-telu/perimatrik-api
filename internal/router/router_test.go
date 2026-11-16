package router

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/config"
	"perimatrik-api-go/internal/handlers"
	"perimatrik-api-go/internal/request"
	"perimatrik-api-go/internal/storage"
	"perimatrik-api-go/internal/validation"
)

func setup(t *testing.T) (http.Handler, string) {
	t.Setenv("APP_URL", "http://api.test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	cfg.StoragePath = dir
	h := &handlers.Handler{Cfg: cfg, Auth: &auth.Service{}, Storage: &storage.Storage{Root: dir, AppURL: cfg.AppURL}}
	return New(cfg, h), dir
}

func do(r http.Handler, method, path string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func body(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("not json: %s", w.Body.String())
	}
	return m
}

func TestRoutesAndErrors(t *testing.T) {
	r, dir := setup(t)

	if w := do(r, "GET", "/", nil); w.Code != 200 || body(t, w)["message"] != "Hello World" {
		t.Fatalf("root: %d %s", w.Code, w.Body)
	}

	w := do(r, "GET", "/api/v1/auth/current-user", nil)
	b := body(t, w)
	if w.Code != 401 || b["message"] != "Unauthenticated." || b["error_code"].(float64) != 1101 || b["success"] != false {
		t.Fatalf("401: %d %s", w.Code, w.Body)
	}

	w = do(r, "GET", "/api/v1/system/users", nil)
	if w.Code != 401 {
		t.Fatalf("admin route without token: %d", w.Code)
	}

	w = do(r, "GET", "/api/v1/does-not-exist", nil)
	if b := body(t, w); w.Code != 404 || b["message"] != "Item not found" || b["error_code"].(float64) != 1105 {
		t.Fatalf("404: %d %s", w.Code, w.Body)
	}

	w = do(r, "DELETE", "/api/v1/content/article", nil)
	if b := body(t, w); w.Code != 405 || b["error_code"].(float64) != 1106 {
		t.Fatalf("405: %d %s", w.Code, w.Body)
	}

	// CORS preflight
	w = do(r, "OPTIONS", "/api/v1/auth/login", map[string]string{
		"Origin": "https://dashboard.perimatrik.org", "Access-Control-Request-Method": "POST",
		"Access-Control-Request-Headers": "authorization,content-type",
	})
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Origin") != "https://dashboard.perimatrik.org" ||
		w.Header().Get("Access-Control-Allow-Credentials") != "true" ||
		w.Header().Get("Access-Control-Allow-Headers") != "authorization,content-type" {
		t.Fatalf("preflight: %d %v", w.Code, w.Header())
	}
	w = do(r, "OPTIONS", "/api/v1/auth/login", map[string]string{"Origin": "https://evil.com", "Access-Control-Request-Method": "POST"})
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("evil origin allowed")
	}
	w = do(r, "OPTIONS", "/api/v1/auth/login", map[string]string{"Origin": "http://foo.localhost:5173", "Access-Control-Request-Method": "POST"})
	if w.Header().Get("Access-Control-Allow-Origin") != "http://foo.localhost:5173" {
		t.Fatal("pattern origin not allowed")
	}

	// image + storage serving, path traversal
	os.MkdirAll(filepath.Join(dir, "galleries"), 0o755)
	os.WriteFile(filepath.Join(dir, "galleries", "a.png"), []byte("\x89PNG\r\n\x1a\nxxxx"), 0o644)
	if w := do(r, "GET", "/api/v1/images/galleries/a.png", nil); w.Code != 200 || w.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("image: %d %v", w.Code, w.Header())
	}
	if w := do(r, "GET", "/storage/galleries/a.png", nil); w.Code != 200 {
		t.Fatalf("storage: %d", w.Code)
	}
	if w := do(r, "GET", "/api/v1/images/../../etc/passwd", nil); w.Code != 404 {
		t.Fatalf("traversal: %d", w.Code)
	}
	if w := do(r, "GET", "/api/v1/images/galleries", nil); w.Code != 404 {
		t.Fatalf("dir: %d", w.Code)
	}
}

func TestVercelFrontendCORS(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("CORS_ALLOWED_ORIGIN_PATTERNS", "")
	r, _ := setup(t)
	for _, origin := range []string{"https://perimatrik.vercel.app", "https://unrelated.vercel.app"} {
		w := do(r, "OPTIONS", "/api/v1/content/article?page=1&limit=12", map[string]string{
			"Origin": origin, "Access-Control-Request-Method": "GET",
			"Access-Control-Request-Headers": "authorization,content-type",
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("preflight status: %d", w.Code)
		}
		if origin != "https://perimatrik.vercel.app" {
			if w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Fatal("unrelated Vercel origin allowed")
			}
			continue
		}
		if w.Header().Get("Access-Control-Allow-Origin") != origin ||
			w.Header().Get("Access-Control-Allow-Methods") != "GET" ||
			w.Header().Get("Access-Control-Allow-Headers") != "authorization,content-type" ||
			w.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Fatalf("frontend preflight headers: %v", w.Header())
		}
		w = do(r, "GET", "/", map[string]string{"Origin": origin})
		if w.Code != http.StatusOK || w.Header().Get("Access-Control-Allow-Origin") != origin {
			t.Fatalf("frontend GET: %d %v", w.Code, w.Header())
		}
	}
}

func TestMultipartInputAndFileRules(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("caption", "  Foto Bersama  ")
	mw.WriteField("latitude", "")
	mw.WriteField("_method", "PUT")
	fw, _ := mw.CreateFormFile("image", "photo.png")
	fw.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"))
	fw2, _ := mw.CreateFormFile("avatar", "evil.png")
	fw2.Write([]byte("<html><script>alert(1)</script></html>"))
	mw.Close()

	req := httptest.NewRequest("POST", "/x?page=2", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	in, err := request.Parse(req)
	if err != nil {
		t.Fatal(err)
	}
	if in.String("caption") != "Foto Bersama" || in.Str("latitude") != nil || !in.Has("latitude") || in.Int("page", 1) != 2 {
		t.Fatalf("parse: caption=%q lat=%v", in.String("caption"), in.Str("latitude"))
	}

	v := validation.New(nil)
	errs := v.Validate(in,
		validation.R("image", "required|file|mimes:jpeg,png,jpg,gif,svg|max:2048"),
		validation.R("avatar", "nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048"),
		validation.R("latitude", "nullable|numeric|between:-90,90"),
	)
	got, _ := json.Marshal(errs)
	if string(got) != `{"avatar":["The avatar field must be a file of type: jpeg, png, jpg, gif, svg."]}` {
		t.Fatalf("errors: %s", got)
	}

	st := &storage.Storage{Root: t.TempDir(), AppURL: "http://api.test"}
	url, err := st.SaveURL(in.File("image"), storage.GalleryDir)
	if err != nil || len(url) != len("http://api.test/storage/galleries/")+30+4 {
		t.Fatalf("save: %v %s", err, url)
	}
}
