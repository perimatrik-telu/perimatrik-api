// Package validation implements the subset of Laravel validation rules used by
// the API, with the same rule syntax ("required|string|max:255") and the same
// English error messages, so clients receive identical `errors` objects.
package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"perimatrik-api-go/internal/request"
	"perimatrik-api-go/internal/support"

	"gorm.io/gorm"
)

// Errors is an ordered map of field => messages (JSON object in rule order).
type Errors struct {
	keys []string
	msgs map[string][]string
}

func NewErrors() *Errors { return &Errors{msgs: map[string][]string{}} }

// Single builds an Errors with one message (ValidationException::withMessages).
func Single(field, msg string) *Errors {
	e := NewErrors()
	e.Add(field, msg)
	return e
}

func (e *Errors) Add(field, msg string) {
	if _, ok := e.msgs[field]; !ok {
		e.keys = append(e.keys, field)
	}
	e.msgs[field] = append(e.msgs[field], msg)
}

func (e *Errors) has(field string) bool { return len(e.msgs[field]) > 0 }

func (e *Errors) Empty() bool { return len(e.keys) == 0 }

// Summary mirrors ValidationException::summarize():
// "first message (and N more errors)".
func (e *Errors) Summary() string {
	if e.Empty() {
		return "The given data was invalid."
	}
	total := 0
	for _, k := range e.keys {
		total += len(e.msgs[k])
	}
	msg := e.msgs[e.keys[0]][0]
	if n := total - 1; n > 0 {
		word := "errors"
		if n == 1 {
			word = "error"
		}
		msg += fmt.Sprintf(" (and %d more %s)", n, word)
	}
	return msg
}

func (e *Errors) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range e.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, _ := json.Marshal(k)
		vb, err := json.Marshal(e.msgs[k])
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// Rule is a field with its pipe-separated rules.
type Rule struct {
	Field string
	Rules []string
}

func R(field, rules string) Rule { return Rule{Field: field, Rules: strings.Split(rules, "|")} }

type Validator struct{ DB *gorm.DB }

func New(db *gorm.DB) *Validator { return &Validator{DB: db} }

// Validate checks data against rules and returns the (possibly empty) errors.
func (v *Validator) Validate(data *request.Input, rules ...Rule) *Errors {
	errs := NewErrors()
	for _, r := range rules {
		v.validateField(data, r, errs)
	}
	return errs
}

func (v *Validator) validateField(data *request.Input, r Rule, errs *Errors) {
	value, present := data.Get(r.Field)
	nullable, required, numeric := false, false, false
	for _, rule := range r.Rules {
		switch rule {
		case "nullable":
			nullable = true
		case "required":
			required = true
		case "numeric":
			numeric = true
		}
	}
	if !present && !required {
		return
	}
	if value == nil && nullable && !required {
		return
	}

	attr := strings.ReplaceAll(r.Field, "_", " ")
	for _, rule := range r.Rules {
		name, param, _ := strings.Cut(rule, ":")
		if (name == "unique" || name == "exists") && errs.has(r.Field) {
			continue // Laravel skips DB rules once the attribute has failed
		}
		msg := v.check(data, r.Field, attr, name, param, value, numeric)
		if msg != "" {
			errs.Add(r.Field, msg)
			if name == "required" {
				return
			}
		}
	}
}

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (v *Validator) check(data *request.Input, field, attr, name, param string, value any, numeric bool) string {
	str, isStr := value.(string)
	file, isFile := value.(*multipart.FileHeader)

	switch name {
	case "nullable", "numeric_marker":
		return ""
	case "required":
		if value == nil {
			return fmt.Sprintf("The %s field is required.", attr)
		}
		if arr, ok := value.([]any); ok && len(arr) == 0 {
			return fmt.Sprintf("The %s field is required.", attr)
		}
	case "string":
		if !isStr {
			return fmt.Sprintf("The %s field must be a string.", attr)
		}
	case "email":
		if !isStr || !validEmail(str) {
			return fmt.Sprintf("The %s field must be a valid email address.", attr)
		}
	case "uuid":
		if s := request.Stringify(value); !uuidRe.MatchString(s) {
			return fmt.Sprintf("The %s field must be a valid UUID.", attr)
		}
	case "url":
		if !isStr || !validURL(str) {
			return fmt.Sprintf("The %s field must be a valid URL.", attr)
		}
	case "date":
		if !isStr {
			return fmt.Sprintf("The %s field must be a valid date.", attr)
		}
		if _, ok := support.ParseDate(str, nil); !ok {
			return fmt.Sprintf("The %s field must be a valid date.", attr)
		}
	case "boolean":
		switch x := value.(type) {
		case bool:
			return ""
		case json.Number:
			if x == "0" || x == "1" {
				return ""
			}
		case string:
			if x == "0" || x == "1" {
				return ""
			}
		}
		return fmt.Sprintf("The %s field must be true or false.", attr)
	case "numeric":
		if _, ok := toNumber(value); !ok {
			return fmt.Sprintf("The %s field must be a number.", attr)
		}
	case "in":
		s := request.Stringify(value)
		for _, opt := range strings.Split(param, ",") {
			if s == opt && value != nil {
				return ""
			}
		}
		return fmt.Sprintf("The selected %s is invalid.", attr)
	case "confirmed":
		other, _ := data.Get(field + "_confirmation")
		if other == nil || request.Stringify(other) != request.Stringify(value) {
			return fmt.Sprintf("The %s field confirmation does not match.", attr)
		}
	case "file":
		if !isFile {
			return fmt.Sprintf("The %s field must be a file.", attr)
		}
	case "mimes":
		if !isFile || !matchesMimes(file, strings.Split(param, ",")) {
			return fmt.Sprintf("The %s field must be a file of type: %s.", attr, strings.ReplaceAll(param, ",", ", "))
		}
	case "max", "min":
		limit, _ := strconv.ParseFloat(param, 64)
		size, kind := sizeOf(value, numeric)
		if (name == "max" && size > limit) || (name == "min" && size < limit) {
			if name == "max" {
				return fmt.Sprintf("The %s field must not be greater than %s%s.", attr, param, kind)
			}
			return fmt.Sprintf("The %s field must be at least %s%s.", attr, param, kind)
		}
	case "between":
		lo, hi, _ := strings.Cut(param, ",")
		min, _ := strconv.ParseFloat(lo, 64)
		max, _ := strconv.ParseFloat(hi, 64)
		size, kind := sizeOf(value, numeric)
		if size < min || size > max {
			return fmt.Sprintf("The %s field must be between %s and %s%s.", attr, lo, hi, kind)
		}
	case "unique":
		args := strings.Split(param, ",")
		table, column := args[0], field
		if len(args) > 1 && args[1] != "" {
			column = args[1]
		}
		q := v.DB.Table(table).Where(quote(column)+" = ?", request.Stringify(value))
		if len(args) > 2 && args[2] != "" && args[2] != "NULL" {
			idCol := "id"
			if len(args) > 3 && args[3] != "" {
				idCol = args[3]
			}
			q = q.Where(quote(idCol)+" <> ?", args[2])
		}
		var n int64
		if err := q.Count(&n).Error; err != nil || n > 0 {
			return fmt.Sprintf("The %s has already been taken.", attr)
		}
	case "exists":
		args := strings.Split(param, ",")
		table, column := args[0], field
		if len(args) > 1 && args[1] != "" {
			column = args[1]
		}
		var n int64
		if err := v.DB.Table(table).Where(quote(column)+" = ?", request.Stringify(value)).Count(&n).Error; err != nil || n == 0 {
			return fmt.Sprintf("The selected %s is invalid.", attr)
		}
	default:
		panic("validation: unsupported rule " + name)
	}
	return ""
}

func quote(col string) string { return "`" + strings.ReplaceAll(col, "`", "") + "`" }

func validEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address == s && strings.Contains(s, "@")
}

func validURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func toNumber(v any) (float64, bool) {
	switch x := v.(type) {
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return f, err == nil
	}
	return 0, false
}

// sizeOf mirrors Validator::getSize: numeric value, file kilobytes, array
// length or string character count. kind is the message suffix.
func sizeOf(v any, numeric bool) (float64, string) {
	if f, ok := v.(*multipart.FileHeader); ok {
		return float64(f.Size) / 1024, " kilobytes"
	}
	if numeric {
		if n, ok := toNumber(v); ok {
			return n, ""
		}
	}
	if arr, ok := v.([]any); ok {
		return float64(len(arr)), " items"
	}
	return float64(utf8.RuneCountInString(request.Stringify(v))), " characters"
}

// matchesMimes checks the file's real content type against extensions,
// like Laravel's `mimes` rule (which guesses the extension from content).
func matchesMimes(fh *multipart.FileHeader, exts []string) bool {
	ext := DetectExtension(fh)
	if ext == "" {
		return false
	}
	for _, e := range exts {
		if e == ext || (ext == "jpg" && e == "jpeg") || (ext == "jpeg" && e == "jpg") {
			return true
		}
	}
	return false
}

// DetectExtension guesses a file extension from the uploaded content.
func DetectExtension(fh *multipart.FileHeader) string {
	f, err := fh.Open()
	if err != nil {
		return ""
	}
	defer f.Close()
	head := make([]byte, 1024)
	n, _ := io.ReadFull(f, head)
	head = head[:n]

	switch http.DetectContentType(head) {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/bmp":
		return "bmp"
	}
	lower := bytes.ToLower(head)
	if bytes.Contains(lower, []byte("<svg")) {
		return "svg"
	}
	return ""
}
