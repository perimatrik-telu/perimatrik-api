// Package support holds small helpers ported from Laravel's Str and PHP.
package support

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const alnum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandomString mirrors Str::random: n cryptographically random [A-Za-z0-9].
func RandomString(n int) string {
	b := make([]byte, n)
	max := big.NewInt(int64(len(alnum)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = alnum[idx.Int64()]
	}
	return string(b)
}

var (
	dashes        = strings.NewReplacer("‐", "-", "‑", "-", "‒", "-", "–", "-", "—", "-", "―", "-", "−", "-")
	slugInvalid   = regexp.MustCompile(`[^-\pL\pN\s]+`)
	slugSeparator = regexp.MustCompile(`[-\s]+`)
)

// Slug mirrors Str::slug($title, '-').
func Slug(title string) string {
	s := toASCII(title)
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, "@", "-at-")
	s = slugInvalid.ReplaceAllString(strings.ToLower(s), "")
	s = slugSeparator.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// toASCII strips diacritics (é -> e) and drops other non-ASCII characters,
// approximating Str::ascii.
func toASCII(s string) string {
	s = dashes.Replace(s)
	var b strings.Builder
	for _, r := range norm.NFKD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if r < unicode.MaxASCII {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var tagRe = regexp.MustCompile(`(?s)<!--.*?-->|<[^>]*>`)

// StripTags approximates PHP strip_tags.
func StripTags(s string) string { return tagRe.ReplaceAllString(s, "") }

// Limit mirrors Str::limit($value, $limit): truncate to n characters + "...".
func Limit(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimRight(string(r[:n]), " \t\n\r\x00\x0B") + "..."
}

// FilterBool mirrors filter_var($v, FILTER_VALIDATE_BOOLEAN).
func FilterBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "on", "yes":
		return true
	}
	return false
}

var dateLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05.000Z07:00",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006-01-02",
	"2006/01/02 15:04:05",
	"2006/01/02",
	time.RFC1123Z,
	time.RFC1123,
}

// ParseDate parses the date formats a frontend may send. Values without a
// zone are interpreted in loc (nil = UTC).
func ParseDate(s string, loc *time.Location) (time.Time, bool) {
	if loc == nil {
		loc = time.UTC
	}
	s = strings.TrimSpace(s)
	for _, layout := range dateLayouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
