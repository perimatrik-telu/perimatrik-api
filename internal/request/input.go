// Package request gives handlers a Laravel-like view of request input: query
// string, JSON, urlencoded and multipart bodies merged into one bag, with
// strings trimmed and empty strings converted to null (Laravel's TrimStrings
// and ConvertEmptyStringsToNull middleware).
package request

import (
	"encoding/json"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

const maxMemory = 32 << 20

var untrimmed = map[string]bool{
	"password":              true,
	"password_confirmation": true,
	"current_password":      true,
}

type Input struct {
	values map[string]any
	files  map[string]*multipart.FileHeader
}

func New() *Input {
	return &Input{values: map[string]any{}, files: map[string]*multipart.FileHeader{}}
}

// Parse reads all input of r. Body values override query values.
func Parse(r *http.Request) (*Input, error) {
	in := New()
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			in.set(k, v[0])
		}
	}

	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch {
	case ct == "application/json" || strings.HasSuffix(ct, "+json"):
		if r.Body == nil || r.ContentLength == 0 {
			break
		}
		var body map[string]any
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		if err := dec.Decode(&body); err != nil {
			return nil, fmt.Errorf("invalid JSON body: %w", err)
		}
		for k, v := range body {
			in.set(k, v)
		}
	case ct == "multipart/form-data":
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			return nil, fmt.Errorf("invalid multipart body: %w", err)
		}
		for k, v := range r.MultipartForm.Value {
			if len(v) > 0 {
				in.set(k, v[0])
			}
		}
		for k, f := range r.MultipartForm.File {
			if len(f) > 0 && f[0].Size > 0 && f[0].Filename != "" {
				in.files[k] = f[0]
				delete(in.values, k)
			}
		}
	case ct == "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			return nil, fmt.Errorf("invalid form body: %w", err)
		}
		for k, v := range r.PostForm {
			if len(v) > 0 {
				in.set(k, v[0])
			}
		}
	}
	return in, nil
}

func (in *Input) set(k string, v any) {
	if s, ok := v.(string); ok {
		if !untrimmed[k] {
			s = strings.TrimSpace(s)
		}
		if s == "" {
			in.values[k] = nil
			return
		}
		v = s
	}
	in.values[k] = v
}

// Set stores a value as if it came from the request.
func (in *Input) Set(k string, v any) { in.set(k, v) }

// SetFile stores an uploaded file (no-op for nil).
func (in *Input) SetFile(k string, f *multipart.FileHeader) {
	if f != nil {
		in.files[k] = f
	}
}

// Get returns the raw value (string, json.Number, bool, nil, []any, map or
// *multipart.FileHeader) and whether the key was present.
func (in *Input) Get(k string) (any, bool) {
	if f, ok := in.files[k]; ok {
		return f, true
	}
	v, ok := in.values[k]
	return v, ok
}

// Has reports whether the key is present (even if null).
func (in *Input) Has(k string) bool { _, ok := in.Get(k); return ok }

// Filled reports whether the key is present and not null.
func (in *Input) Filled(k string) bool {
	v, ok := in.Get(k)
	return ok && v != nil
}

func (in *Input) HasFile(k string) bool { return in.files[k] != nil }

func (in *Input) File(k string) *multipart.FileHeader { return in.files[k] }

// Str returns the value as a string pointer, nil when absent or null.
func (in *Input) Str(k string) *string {
	v, ok := in.values[k]
	if !ok || v == nil {
		return nil
	}
	s := Stringify(v)
	return &s
}

// String returns the value as a string, "" when absent or null.
func (in *Input) String(k string) string {
	if p := in.Str(k); p != nil {
		return *p
	}
	return ""
}

// StringOr returns the value or def when absent or null.
func (in *Input) StringOr(k, def string) string {
	if p := in.Str(k); p != nil {
		return *p
	}
	return def
}

// Int parses an integer value, returning def when absent or invalid.
func (in *Input) Int(k string, def int) int {
	p := in.Str(k)
	if p == nil {
		return def
	}
	n, err := strconv.Atoi(*p)
	if err != nil {
		if f, ferr := strconv.ParseFloat(*p, 64); ferr == nil {
			return int(f)
		}
		return def
	}
	return n
}

// Only returns a copy containing only the given keys.
func (in *Input) Only(keys ...string) *Input {
	out := New()
	for _, k := range keys {
		if v, ok := in.values[k]; ok {
			out.values[k] = v
		}
		if f, ok := in.files[k]; ok {
			out.files[k] = f
		}
	}
	return out
}

// Except returns a copy without the given keys.
func (in *Input) Except(keys ...string) *Input {
	skip := map[string]bool{}
	for _, k := range keys {
		skip[k] = true
	}
	out := New()
	for k, v := range in.values {
		if !skip[k] {
			out.values[k] = v
		}
	}
	for k, f := range in.files {
		if !skip[k] {
			out.files[k] = f
		}
	}
	return out
}

// Stringify converts a scalar input value to its PHP-like string form.
func Stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case json.Number:
		return x.String()
	case bool:
		if x {
			return "1"
		}
		return ""
	default:
		return fmt.Sprint(x)
	}
}
