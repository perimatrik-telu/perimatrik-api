// Package models maps the existing Laravel MySQL tables. JSON output mirrors
// the Eloquent serialisation (hidden fields, appended accessors, relations).
package models

import (
	"encoding/json"
	"strings"

	"gorm.io/gorm"
)

// ---------------------------------------------------------------- User

type User struct {
	ID              string  `gorm:"primaryKey" json:"id"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	Password        string  `json:"-"`
	RememberToken   *string `json:"-"`
	EmailVerifiedAt *Time   `json:"-"`
	Role            string  `gorm:"default:member" json:"role"`
	Avatar          *string `json:"avatar"`
	CreatedAt       Time    `json:"-"`
	UpdatedAt       Time    `json:"-"`

	Member *Member `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) BeforeCreate(*gorm.DB) error { setID(&u.ID); return nil }

// UserWithMember serialises a user together with its (possibly null) member,
// like `User::with('member.angkatan')`.
type UserWithMember struct {
	User
	Member *Member `json:"member"`
}

func (u User) WithMember() UserWithMember { return UserWithMember{User: u, Member: u.Member} }

// ---------------------------------------------------------------- Article / News

type Article struct {
	ID          string  `gorm:"primaryKey" json:"id"`
	Slug        string  `json:"slug"`
	Image       string  `json:"image"`
	Title       string  `json:"title"`
	Body        string  `json:"body"`
	Views       uint    `json:"views"`
	Status      string  `gorm:"default:draft" json:"status"`
	PublishedAt *Time   `json:"published_at"`
	UserID      *string `json:"user_id"`
	CreatedAt   Time    `json:"created_at"`
	UpdatedAt   Time    `json:"updated_at"`

	User *User `json:"user"`
}

func (a *Article) BeforeCreate(*gorm.DB) error { setID(&a.ID); return nil }

func (a Article) MarshalJSON() ([]byte, error) {
	type alias Article
	return json.Marshal(struct {
		alias
		postAppends
	}{alias(a), newPostAppends(a.User, a.PublishedAt, a.CreatedAt)})
}

type News struct {
	ID          string  `gorm:"primaryKey" json:"id"`
	Slug        string  `json:"slug"`
	Image       *string `json:"image"`
	Title       string  `json:"title"`
	Body        string  `json:"body"`
	Views       uint    `json:"views"`
	Status      string  `gorm:"default:draft" json:"status"`
	PublishedAt *Time   `json:"published_at"`
	UserID      *string `json:"user_id"`
	CreatedAt   Time    `json:"created_at"`
	UpdatedAt   Time    `json:"updated_at"`

	User *User `json:"user"`
}

func (News) TableName() string { return "news" }

func (n *News) BeforeCreate(*gorm.DB) error { setID(&n.ID); return nil }

func (n News) MarshalJSON() ([]byte, error) {
	type alias News
	return json.Marshal(struct {
		alias
		postAppends
	}{alias(n), newPostAppends(n.User, n.PublishedAt, n.CreatedAt)})
}

// postAppends are the Eloquent $appends of Article and News.
type postAppends struct {
	Author    *string  `json:"author"`
	PostingAt string   `json:"postingAt"`
	Tags      []string `json:"tags"`
}

func newPostAppends(u *User, publishedAt *Time, createdAt Time) postAppends {
	p := postAppends{Tags: []string{}}
	if u != nil {
		p.Author = &u.Name
	}
	t := createdAt.Std()
	if publishedAt != nil {
		t = publishedAt.Std()
	}
	p.PostingAt = t.In(Location).Format("2006-01-02T15:04:05-07:00")
	return p
}

// ---------------------------------------------------------------- Gallery

type Gallery struct {
	ID        string  `gorm:"primaryKey" json:"id"`
	Image     string  `json:"image"`
	Caption   *string `json:"caption"`
	Slug      *string `json:"slug"`
	UserID    *string `json:"user_id"`
	CreatedAt Time    `json:"created_at"`
	UpdatedAt Time    `json:"updated_at"`
	Latitude  *string `json:"latitude"`
	Longitude *string `json:"longitude"`

	User *User `json:"user"`
}

func (Gallery) TableName() string { return "galleries" }

func (g *Gallery) BeforeCreate(*gorm.DB) error { setID(&g.ID); return nil }

func (g Gallery) MarshalJSON() ([]byte, error) {
	type alias Gallery
	return json.Marshal(struct {
		alias
		ImageURL string `json:"image_url"`
	}{alias(g), g.ImageURL()})
}

// ImageURL mirrors Gallery::getImageUrlAttribute: files on the public disk are
// served through the /api/v1/images/{path} route.
func (g Gallery) ImageURL() string {
	if g.Image == "" {
		return ""
	}
	if i := strings.Index(g.Image, "/storage/"); i >= 0 {
		return AppURL + "/api/v1/images/" + g.Image[i+len("/storage/"):]
	}
	return g.Image
}

// ---------------------------------------------------------------- Angkatan / Member

type Angkatan struct {
	ID           string  `gorm:"primaryKey" json:"id"`
	NamaAngkatan string  `json:"nama_angkatan"`
	Year         *string `json:"year"`
	CreatedAt    Time    `json:"created_at"`
	UpdatedAt    Time    `json:"updated_at"`
}

func (Angkatan) TableName() string { return "angkatans" }

func (a *Angkatan) BeforeCreate(*gorm.DB) error { setID(&a.ID); return nil }

type Member struct {
	ID           string  `gorm:"primaryKey" json:"id"`
	UserID       *string `json:"user_id"`
	NoAnggota    string  `json:"no_anggota"`
	NamaLengkap  string  `json:"nama_lengkap"`
	NamaLapangan *string `json:"nama_lapangan"`
	Domisili     *string `json:"domisili"`
	Kontak       *string `json:"kontak"`
	Avatar       *string `json:"avatar"`
	AngkatanID   *string `json:"angkatan_id"`
	CreatedAt    Time    `json:"created_at"`
	UpdatedAt    Time    `json:"updated_at"`
	Instagram    *string `json:"instagram"`
	Facebook     *string `json:"facebook"`
	Linkedin     *string `json:"linkedin"`
	Tiktok       *string `json:"tiktok"`
	Youtube      *string `json:"youtube"`

	Angkatan *Angkatan `json:"angkatan"`
}

func (m *Member) BeforeCreate(*gorm.DB) error { setID(&m.ID); return nil }

// MemberFillable lists the mass-assignable columns of Member ($fillable).
var MemberFillable = []string{
	"no_anggota", "nama_lengkap", "nama_lapangan", "domisili", "kontak", "avatar",
	"angkatan_id", "user_id", "instagram", "facebook", "linkedin", "tiktok", "youtube",
}

// ---------------------------------------------------------------- Setting

type Setting struct {
	ID        string  `gorm:"primaryKey" json:"id"`
	Key       string  `json:"key"`
	Value     *string `json:"value"`
	CreatedAt Time    `json:"created_at"`
	UpdatedAt Time    `json:"updated_at"`
}

func (s *Setting) BeforeCreate(*gorm.DB) error { setID(&s.ID); return nil }

// ---------------------------------------------------------------- Comment

type Comment struct {
	ID        string  `gorm:"primaryKey" json:"id"`
	Body      string  `json:"body"`
	UserID    string  `json:"user_id"`
	ArticleID *string `json:"article_id"`
	NewsID    *string `json:"news_id"`
	ParentID  *string `json:"parent_id"`
	CreatedAt Time    `json:"created_at"`
	UpdatedAt Time    `json:"updated_at"`

	User    *User      `json:"user"`
	Replies []*Comment `gorm:"-" json:"replies"`
}

func (c *Comment) BeforeCreate(*gorm.DB) error { setID(&c.ID); return nil }

// ---------------------------------------------------------------- Sanctum token

type PersonalAccessToken struct {
	ID            uint64 `gorm:"primaryKey"`
	TokenableType string
	TokenableID   string
	Name          string
	Token         string
	Abilities     *string
	LastUsedAt    *Time
	ExpiresAt     *Time
	CreatedAt     Time
	UpdatedAt     Time
}

func setID(id *string) {
	if *id == "" {
		*id = NewUUID()
	}
}
