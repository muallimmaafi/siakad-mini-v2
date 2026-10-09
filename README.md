# SIAKAD Mini API

RESTful API back end untuk SIAKAD Mini, layanan akademik sederhana yang mengelola
data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS). Dibuat untuk UTS
Pemrograman Backend Lanjut (Praktikum).

## Tech Stack

- Go dan Fiber v2
- PostgreSQL (driver pgx, connection pool, SQL berparameter)
- JWT (`golang-jwt/jwt/v5`) untuk autentikasi
- bcrypt untuk hash password
- slog dengan rotasi log (lumberjack)

## Fitur

- Dua peran: `admin` dan `mahasiswa`
- Autentikasi Bearer token di semua endpoint kecuali login
- Rate limit login: maksimal 5 kali gagal per menit per IP (429)
- Soft delete mahasiswa (tidak muncul di daftar dan tidak bisa login)
- Pengambilan mata kuliah dalam satu transaction dengan row locking
  (`SELECT ... FOR UPDATE`)
- Format response JSON seragam, error 500 tanpa stack trace

## Struktur Folder

```
app/
  model/        struktur data dan request
  repository/   akses database (SQL berparameter)
  service/      validasi, aturan bisnis, dan handler
config/         env, logger, dan perakitan aplikasi
database/       koneksi PostgreSQL (connection pool)
helper/         response, JWT, bcrypt, request
middleware/     auth, role, rate limit, logger
migrations/     file SQL migrasi dan seeder
route/          pendaftaran route
main.go         entry point
```

## Cara Menjalankan

Prasyarat: Go, PostgreSQL, Git.

```bash
git clone https://github.com/muallimmaafi/siakad-mini-v2.git
cd siakad-mini-v2

# 1. Buat database
psql -U postgres -c "CREATE DATABASE siakad_mini_pgx;"

# 2. Jalankan migrasi dan seeder
psql -U postgres -d siakad_mini_pgx -f migrations/001_create_tables.sql
psql -U postgres -d siakad_mini_pgx -f migrations/002_seed.sql

# 3. Siapkan konfigurasi, lalu isi DB_PASSWORD dan JWT_SECRET (minimal 32 karakter)
cp .env.example .env        # PowerShell: Copy-Item .env.example .env

# 4. Jalankan API
go run .
```

API berjalan di `http://localhost:3000`. Pengecekan cepat:
`GET http://localhost:3000/api/v1/health`.

Menjalankan tes unit: `go test ./...`

## Data Seeder

| Data | Jumlah | Keterangan |
|---|---|---|
| Admin | 1 | `admin@siakad.test` / `admin12345` |
| Mahasiswa | 20 | `mhs01@siakad.test` s.d. `mhs20@siakad.test`, NIM `187221000001` s.d. `187221000020`, **password awal = NIM** |
| Mata kuliah | 10 | `TI101` s.d. `TI502` (TI502 berkuota 2 untuk uji kuota penuh) |

Seeder aman dijalankan berulang kali (`ON CONFLICT DO NOTHING`).

## Daftar Endpoint

Base URL: `/api/v1`

| No | Method | Endpoint | Akses | Status sukses |
|---|---|---|---|---|
| 1 | POST | `/auth/login` | Publik | 200 |
| 2 | GET | `/auth/me` | Semua role | 200 |
| 3 | GET | `/students` | Admin | 200 |
| 4 | POST | `/students` | Admin | 201 |
| 5 | GET | `/students/{id}` | Admin, mahasiswa (data sendiri) | 200 |
| 6 | PUT | `/students/{id}` | Admin | 200 |
| 7 | DELETE | `/students/{id}` | Admin | 204 |
| 8 | GET | `/courses` | Semua role | 200 |
| 9 | POST | `/enrollments` | Mahasiswa | 201 |
| 10 | DELETE | `/enrollments/{id}` | Mahasiswa (milik sendiri) | 204 |

Query parameter:

- `GET /students`: `page`, `per_page` (maks. 50), `prodi`, `angkatan`, `search`
  (nim atau nama), `sort` (`nama` atau `-ipk_terakhir`)
- `GET /students/{id}`: `tahun_akademik` (opsional, menyaring KRS dan total SKS)
- `GET /courses`: `semester`, `search` (kode atau nama), `available=true`,
  `tahun_akademik` (opsional, kuota dihitung per tahun akademik)

## Business Rule

1. Batas SKS per tahun akademik berdasarkan IPK terakhir: IPK >= 3,00 maksimal 24
   SKS; 2,50 sampai 2,99 maksimal 21 SKS; di bawah 2,50 maksimal 18 SKS
   (melebihi batas: 422, pesan menyebut sisa SKS).
2. Mata kuliah yang sama tidak dapat diambil dua kali pada tahun akademik yang
   sama (409).
3. Mata kuliah yang kuotanya penuh tidak dapat diambil (422).
4. Mahasiswa hanya dapat mengakses dan mengubah KRS miliknya sendiri (403).

## Format Response

Sukses:

```json
{
  "success": true,
  "message": "Data mahasiswa berhasil diambil",
  "data": [],
  "meta": { "current_page": 1, "per_page": 10, "total": 20, "last_page": 2 }
}
```

Error validasi (422):

```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": { "nim": ["NIM sudah terdaftar"] }
}
```

Status code yang dipakai: 200, 201, 204, 401, 403, 404, 409, 415, 422, 429, 500.

## Catatan Desain

- Kuota dan batas SKS dihitung per `tahun_akademik`, karena mata kuliah yang sama
  dibuka kembali setiap semester. `terisi` dan `sisa_kuota` dihitung dari tabel
  `enrollments`, sehingga kuota kembali otomatis saat KRS dibatalkan.
- `POST /enrollments` memakai satu transaction: baris mahasiswa dan baris mata
  kuliah dikunci dengan `FOR UPDATE` sebelum cek duplikasi, kuota, dan batas SKS,
  dengan urutan kunci yang selalu sama untuk mencegah deadlock.
- `POST /students` membuat `users` dan `students` dalam satu transaction;
  password awal adalah NIM yang di-hash.
- Pesan login dibuat sama untuk email tidak terdaftar dan password salah, dengan
  dummy hash agar waktu respons mirip.
- Rate limiter login menyimpan hitungan di memori server dan membatasi per IP.

## Pembuat

Muallim Maafi Ahmad, NIM 434241122, TI-C4, D4 Teknik Informatika, Universitas Airlangga.