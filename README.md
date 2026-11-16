# perimatrik-api-go

Port Go (Gin + GORM) dari `perimatrik-api` (Laravel). Bisa langsung menggantikan versi Laravel:

- URL `/api/v1/...`, method, dan role per endpoint sama persis.
- Format response sama: `{ success, message, data, error_code?, meta?, errors? }`, termasuk pesan validasi Laravel (`The email field is required.` dan seterusnya).
- Memakai database MySQL yang sama, tanpa migrasi data.
- Token **Sanctum tetap valid** (tabel `personal_access_tokens`, format `id|token`, SHA-256), jadi user yang sedang login tidak ter-logout.
- Password bcrypt `$2y$` kompatibel dua arah dengan Laravel.
- File upload memakai struktur folder yang sama (`articles/`, `news/`, `galleries/`, `member-avatars/`, `avatars/`) dan URL `{APP_URL}/storage/...`.

## Menjalankan

```bash
cp .env.example .env     # isi DB_*, APP_URL, STORAGE_PATH
make run                 # atau: go run ./cmd/server
make test
make build               # binary di bin/perimatrik-api
```

Untuk database baru (tanpa Laravel): `mysql perimatrik < database/schema.sql`.

### TiDB Cloud / Vercel

TiDB Cloud mewajibkan koneksi TLS. `DB_TLS=auto` (default) mengaktifkan TLS
dengan verifikasi sertifikat dan hostname untuk `DB_HOST` berakhiran
`.tidbcloud.com`. MySQL lokal tetap memakai koneksi tanpa TLS.

Di Vercel, isi `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, dan
`DB_PASSWORD` sesuai detail koneksi TiDB, lalu set `DB_TLS=true` untuk
mewajibkan TLS (juga untuk hostname kustom). Redeploy setelah mengubah
environment variables atau kode. Nilai `DB_TLS` yang didukung: `auto`, `true`,
dan `false`.

### GitHub Actions

Workflow `.github/workflows/ci-cd.yml` mengikuti pola `perimatrik-web-react`:

- Push ke `main` dan pull request: vet, test, build, lalu deployment preview.
- Push ke `production`: pemeriksaan yang sama, lalu deployment production.
- Bisa dijalankan manual lewat tab Actions. Deployment PR dari fork dilewati.
- URL preview ditambahkan ke komentar PR.

Tambahkan repository secrets di **Settings → Secrets and variables → Actions**:

- `VERCEL_TOKEN`: token Vercel yang punya akses ke project API.
- `VERCEL_ORG_ID`: `team_xeSRNm96pfc3wvNC7jb3WPHW`.
- `VERCEL_PROJECT_ID`: `prj_tCEkI2GcmVv7fehsbKOn2PFv1Z5w` (project API).

Atur environment variables aplikasi di Vercel untuk **Preview** dan
**Production**, termasuk `DB_TLS=true` untuk TiDB dan `APP_URL` yang sesuai.
Workflow mengambil konfigurasi lewat `vercel pull`, kemudian menjalankan
`vercel build` dan `vercel deploy --prebuilt`. `vercel.json` menonaktifkan
deployment dari integrasi Git agar deployment dilakukan oleh Actions.

### Migrasi dari Laravel

1. Salin nilai `DB_*`, `APP_URL`, `FRONTEND_URL`, dan `BCRYPT_ROUNDS` dari `.env` Laravel.
2. Set `STORAGE_PATH` ke `perimatrik-api/storage/app/public` (atau salin isinya) supaya file lama tetap bisa diakses.
3. `APP_URL` harus sama dengan sebelumnya, karena URL gambar yang tersimpan di DB memakai domain itu.
4. Arahkan reverse proxy (nginx) ke port Go. Route `/storage/*` sekarang disajikan langsung oleh aplikasi; tidak perlu symlink lagi.

## Struktur

```
cmd/server/            entrypoint
internal/config/       env (.env) loader
internal/database/     koneksi GORM MySQL
internal/models/       tabel + serialisasi JSON ala Eloquent
internal/auth/         token Sanctum, bcrypt, middleware auth & role
internal/request/      input parser (query/JSON/form/multipart, trim + empty→null)
internal/validation/   rule ala Laravel ("required|string|max:255")
internal/response/     envelope response + error code
internal/storage/      disk "public" (upload, URL, hapus)
internal/handlers/     satu file per controller Laravel
internal/export/       export anggota xlsx/csv
internal/router/       route, CORS, 404/405/500
database/schema.sql    skema MySQL
```

## Perbedaan kecil dari versi Laravel

- `PUT/POST /content/gallery/{id}` dan `DELETE` untuk id yang tidak ada sekarang mengembalikan 404. Sebelumnya 500 karena bug `ResponseMessage::NOT_ITEM`.
- Beberapa response create/update menyertakan relasi lengkap (`user`, `angkatan`, `replies`). Isinya superset dari sebelumnya.
- Ekstensi file upload yang tidak dikenal diganti sesuai isi file (mencegah file `.html` ter-upload).

## Test e2e (perbandingan dengan Laravel)

`e2e/run.py` menjalankan sekitar 140 request yang sama ke server Laravel dan Go, lalu membandingkan response-nya. Siapkan dua database uji dengan isi seed yang sama (`e2e/seed.sql`), jalankan kedua server ke DB masing-masing, lalu:

```bash
python3 e2e/run.py http://127.0.0.1:8101 http://127.0.0.1:8102 /tmp/hasil.json
```

Hasil terakhir: 126/139 identik. Sisanya hanya field tambahan di Go dan URL gallery yang memakai `APP_URL` alih-alih host request (di produksi keduanya sama-sama `https://api.perimatrik.org`).
