package support

import "testing"

// Expected values were produced by Laravel's Str::slug / Str::limit.
func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Halo Dunia!":                       "halo-dunia",
		"  Kegiatan Diklat 2026 @ Bandung ": "kegiatan-diklat-2026-at-bandung",
		"Café déjà vu":                      "cafe-deja-vu",
		"foo_bar--baz":                      "foo-bar-baz",
		"Ünïcödé & Co.":                     "unicode-co",
		"100% Lulus: Angkatan XII":          "100-lulus-angkatan-xii",
		"a—b–c":                             "a-b-c",
		"Pelantikan Anggota (Baru)":         "pelantikan-anggota-baru",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLimit(t *testing.T) {
	if got := Limit(StripTags("<p>Hi <b>there</b></p>"), 150); got != "Hi there" {
		t.Errorf("got %q", got)
	}
	if got := Limit("abcdef", 3); got != "abc..." {
		t.Errorf("got %q", got)
	}
}

func TestRandomString(t *testing.T) {
	if s := RandomString(40); len(s) != 40 {
		t.Fatalf("len %d", len(s))
	}
}
