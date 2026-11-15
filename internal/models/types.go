package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Location is the application timezone (APP_TIMEZONE). Set once at startup.
var Location = time.UTC

// AppURL is the public base URL of the API. Set once at startup.
var AppURL = ""

// UserMorphType is the tokenable_type Laravel Sanctum stores for users.
const UserMorphType = `App\Models\User`

// Time is a database timestamp that serialises to JSON exactly like Laravel
// (e.g. "2026-01-02T03:04:05.000000Z").
type Time time.Time

func Now() Time { return Time(time.Now().In(Location)) }

func (t Time) Std() time.Time { return time.Time(t) }

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).UTC().Format("2006-01-02T15:04:05.000000Z"))
}

func (t *Time) Scan(v any) error {
	switch x := v.(type) {
	case nil:
		*t = Time{}
	case time.Time:
		*t = Time(x)
	case []byte:
		return t.parse(string(x))
	case string:
		return t.parse(x)
	default:
		return fmt.Errorf("models.Time: cannot scan %T", v)
	}
	return nil
}

func (t *Time) parse(s string) error {
	p, err := time.ParseInLocation("2006-01-02 15:04:05", s, Location)
	if err != nil {
		return err
	}
	*t = Time(p)
	return nil
}

func (t Time) Value() (driver.Value, error) {
	return time.Time(t).In(Location), nil
}

// TimePtr converts a time.Time into a *Time.
func TimePtr(t time.Time) *Time {
	v := Time(t.In(Location))
	return &v
}

// NewUUID returns a time-ordered UUID (v7), like Laravel's HasUuids.
func NewUUID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return id.String()
}

// Str returns a pointer to s.
func Str(s string) *string { return &s }
