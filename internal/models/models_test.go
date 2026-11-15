package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestTimeJSON(t *testing.T) {
	b, _ := json.Marshal(Time(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)))
	if string(b) != `"2026-01-02T03:04:05.000000Z"` {
		t.Fatalf("got %s", b)
	}
}

func TestArticleJSONAppends(t *testing.T) {
	pub := Time(time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC))
	a := Article{ID: "a1", Title: "T", PublishedAt: &pub, User: &User{ID: "u1", Name: "Budi", Password: "x"}}
	b, _ := json.Marshal(a)
	s := string(b)
	for _, want := range []string{`"author":"Budi"`, `"postingAt":"2026-03-01T10:00:00+00:00"`, `"tags":[]`, `"user":{"id":"u1"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
	if strings.Contains(s, "password") || strings.Contains(s, `"created_at":"0001`) && false {
		t.Errorf("leaked hidden field: %s", s)
	}
}

func TestUserHiddenFields(t *testing.T) {
	u := User{ID: "u1", Name: "A", Email: "a@b.c", Password: "hash", Role: "admin"}
	b, _ := json.Marshal(u.WithMember())
	if string(b) != `{"id":"u1","name":"A","email":"a@b.c","role":"admin","avatar":null,"member":null}` {
		t.Fatalf("got %s", b)
	}
}

func TestGalleryImageURL(t *testing.T) {
	AppURL = "https://api.example.org"
	g := Gallery{Image: "https://api.example.org/storage/galleries/x.jpg"}
	if g.ImageURL() != "https://api.example.org/api/v1/images/galleries/x.jpg" {
		t.Fatal(g.ImageURL())
	}
}

func dryRun(t *testing.T) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "u:p@tcp(127.0.0.1:1)/x?parseTime=true", SkipInitializeWithVersion: true}),
		&gorm.Config{DryRun: true, SkipDefaultTransaction: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestTimestampsAndUUIDOnCreate(t *testing.T) {
	a := Angkatan{NamaAngkatan: "X", Year: Str("2020")}
	sql := dryRun(t).Create(&a).Statement.SQL.String()
	if !strings.Contains(sql, "`created_at`") || !strings.Contains(sql, "`updated_at`") {
		t.Fatalf("timestamps missing: %s", sql)
	}
	if a.ID == "" || time.Time(a.CreatedAt).IsZero() || time.Time(a.UpdatedAt).IsZero() {
		t.Fatalf("id/timestamps not set: %+v", a)
	}
}

func TestTimestampsOnUpdates(t *testing.T) {
	sql := dryRun(t).Model(&Member{ID: "m1"}).Updates(map[string]any{"kontak": "1"}).Statement.SQL.String()
	if !strings.Contains(sql, "`updated_at`") {
		t.Fatalf("updated_at missing: %s", sql)
	}
	g := Gallery{ID: "g1", Image: "x"}
	sql = dryRun(t).Select("*").Omit("created_at").Save(&g).Statement.SQL.String()
	if !strings.Contains(sql, "`updated_at`") || strings.Contains(sql, "`created_at`") || !strings.HasPrefix(sql, "UPDATE") {
		t.Fatalf("bad save sql: %s", sql)
	}
}
