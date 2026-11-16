# Cloudinary

Cloudinary Free menyediakan penyimpanan gambar dan CDN tanpa kartu kredit saat
pendaftaran. Pemakaian tetap dibatasi kredit gabungan untuk storage, bandwidth,
dan transformasi. Lihat [harga resmi](https://cloudinary.com/pricing).

API menggunakan SDK Go resmi untuk upload bertanda tangan dan penghapusan.
Gambar artikel, galeri, dan avatar yang diunggah melalui alur ini bersifat publik.
Browser tidak menerima API secret. Tidak perlu bucket R2 atau akun AWS.

## 1. Buat akun dan ambil kredensial

1. Daftar [Cloudinary Free](https://cloudinary.com/users/register_free).
2. Buka Cloudinary Console dan pilih product environment yang akan dipakai.
3. Buka halaman **API Keys** (di Console / Settings, sesuai tampilan akun).
4. Catat **Cloud Name**, **API Key**, dan **API Secret**. Jangan menyalin hanya
   nama akun atau product environment sebagai Cloud Name.
5. Tidak perlu membuat unsigned upload preset; API melakukan upload bertanda tangan.

## 2. Environment API Vercel

Di project **perimatrik-api → Settings → Environment Variables** untuk Production:

```dotenv
STORAGE_DRIVER=cloudinary
CLOUDINARY_CLOUD_NAME=<Cloud Name>
CLOUDINARY_API_KEY=<API Key>
CLOUDINARY_API_SECRET=<API Secret>
CLOUDINARY_FOLDER=perimatrik
APP_URL=https://perimatrik-api.vercel.app
```

Untuk Preview, set `CLOUDINARY_FOLDER=perimatrik-preview` agar percobaan upload
dan hapus terpisah dari Production. Nama folder hanya boleh berisi huruf, angka,
underscore, dan tanda hubung. Folder/prefix otomatis digunakan saat upload;
tidak perlu membuatnya secara manual. Upload disimpan di bawah public ID seperti
`perimatrik/articles/<nama-file-yang-dienkode>`.

Kredensial hanya disimpan di environment server, jangan memakai `VITE_*` atau
menaruh API secret di React/Git. Variabel `R2_*` tidak lagi digunakan.
Konfigurasi Cloudinary yang tidak lengkap membuat startup gagal dengan nama
variabel yang perlu diisi; tidak ada fallback ke disk lokal ketika upload gagal.
Gunakan `STORAGE_DRIVER=local` untuk development tanpa Cloudinary.

## 3. Migrasi 56 file backup

File backup sudah dipulihkan ke `storage/app/public`: 15 artikel dan 41 galeri.
Isi `CLOUDINARY_*` yang sama di `.env` lokal API (diabaikan Git), lalu dari
folder `perimatrik-api` jalankan dry run:

```bash
go run ./cmd/migrate-media --source ./storage/app/public
```

Setelah daftar sesuai, lakukan upload:

```bash
go run ./cmd/migrate-media --source ./storage/app/public --apply
```

Perintah tidak mengakses database dan tidak menghapus file lokal. Public ID
dihitung secara deterministik dari folder dan nama file asli. Nama file,
termasuk empat file lama berakhiran titik, dienkode agar kompatibel dengan
aturan public ID Cloudinary. Route API tetap memakai nama file asli.
Cloudinary mendeteksi format gambar dari isinya.

Upload menggunakan `overwrite=false`, sehingga aman dilanjutkan dengan
menjalankan perintah yang sama: aset yang sudah ada tidak ditimpa. Jumlah
"Processed" mencakup aset baru maupun yang sudah ada, bukan jumlah upload baru.
Hanya folder media aplikasi yang diproses; symlink ditolak.

## 4. Referensi URL database lama

Jika data artikel/galeri belum ada di TiDB, pulihkan dump database lama terlebih
dahulu. Migrasi file tidak membuat baris artikel atau galeri.

Dump lama memakai `https://api.perimatrik.org/storage/...`. Setelah semua file
diunggah dan API terbaru di-deploy, gunakan `database/restore-media-urls.sql`
di database aplikasi TiDB. Jalankan SET dan SELECT dulu untuk memeriksa jumlah
baris yang cocok, baru jalankan blok transaksi UPDATE. Script tidak dijalankan
otomatis oleh aplikasi atau migrator.

URL baru tetap `https://perimatrik-api.vercel.app/storage/...`. API memberi
redirect HTTP 302 ke `https://res.cloudinary.com/...`, lalu browser mengambil
gambar langsung dari CDN. `/api/v1/images/...` menggunakan cara yang sama.
Script hanya mengganti prefix URL pada kolom gambar/avatar. Gambar yang
tertanam di body HTML perlu diperiksa terpisah.

## 5. Deploy dan cek

1. Push kode API, lalu jalankan **Actions → CI/CD → Run workflow → production**.
2. Jalankan migrasi upload setelah kredensial lokal tersedia.
3. Cek Cloudinary Media Library: backup awal menghasilkan 56 aset.
4. Buka satu URL `/storage/articles/<nama-asli>` dan pastikan redirect ke CDN
   menghasilkan gambar. Aset yang belum diunggah akan 404 di CDN.
5. Coba upload foto dari dashboard dan reload halaman. Periksa aset di Cloudinary.
6. Hapus foto uji melalui dashboard. API menghapus public ID yang sama dan meminta
   invalidasi cache CDN; salinan cache mungkin belum langsung hilang.

Jangan mengganti `CLOUDINARY_CLOUD_NAME` atau `CLOUDINARY_FOLDER` setelah migrasi
tanpa memindahkan aset, karena keduanya menentukan lokasi gambar.

Referensi: [SDK Go](https://cloudinary.com/documentation/go_integration),
[upload Go](https://cloudinary.com/documentation/go_image_and_video_upload),
[paket Free](https://cloudinary.com/documentation/billing_and_plans).
