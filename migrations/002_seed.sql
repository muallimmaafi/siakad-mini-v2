CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1 admin
INSERT INTO users (email, password, role)
VALUES ('admin@siakad.test', crypt('admin12345', gen_salt('bf', 10)), 'admin')
ON CONFLICT (email) DO NOTHING;

-- 20 mahasiswa: akun (password awal = NIM, di-hash bcrypt) + data mahasiswa
WITH data (urut, nama, prodi, angkatan, ipk) AS (
    VALUES
    (1,  'Rina Putri',       'Sistem Informasi',   2022, 3.45),
    (2,  'Budi Santoso',     'Teknik Informatika', 2022, 3.10),
    (3,  'Siti Aminah',      'Teknik Informatika', 2021, 3.80),
    (4,  'Ahmad Fauzi',      'Sistem Informasi',   2023, 2.75),
    (5,  'Dewi Lestari',     'Teknik Informatika', 2023, 2.60),
    (6,  'Eko Prasetyo',     'Sistem Informasi',   2021, 3.25),
    (7,  'Fitri Handayani',  'Teknik Informatika', 2024, 2.30),
    (8,  'Gilang Ramadhan',  'Sistem Informasi',   2022, 2.90),
    (9,  'Hana Safitri',     'Teknik Informatika', 2023, 3.55),
    (10, 'Indra Kurniawan',  'Sistem Informasi',   2024, 2.45),
    (11, 'Joko Widodo',      'Teknik Informatika', 2021, 3.00),
    (12, 'Kartika Sari',     'Sistem Informasi',   2022, 3.70),
    (13, 'Lukman Hakim',     'Teknik Informatika', 2023, 2.85),
    (14, 'Maya Anggraini',   'Sistem Informasi',   2024, 2.20),
    (15, 'Naufal Rizky',     'Teknik Informatika', 2022, 3.35),
    (16, 'Oktaviani Putri',  'Sistem Informasi',   2021, 2.55),
    (17, 'Putra Mahendra',   'Teknik Informatika', 2023, 3.15),
    (18, 'Qonita Azzahra',   'Sistem Informasi',   2024, 3.90),
    (19, 'Rizal Firmansyah', 'Teknik Informatika', 2022, 2.00),
    (20, 'Sekar Ayu',        'Sistem Informasi',   2023, 3.40)
),
akun AS (
    INSERT INTO users (email, password, role)
    SELECT format('mhs%s@siakad.test', lpad(urut::text, 2, '0')),
           crypt(format('18722100%s', lpad(urut::text, 4, '0')), gen_salt('bf', 10)),
           'mahasiswa'
    FROM data
    ON CONFLICT (email) DO NOTHING
    RETURNING id, email
)
INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT a.id,
       format('18722100%s', lpad(d.urut::text, 4, '0')),
       d.nama, d.prodi, d.angkatan, d.ipk
FROM data d
JOIN akun a ON a.email = format('mhs%s@siakad.test', lpad(d.urut::text, 2, '0'))
ON CONFLICT (nim) DO NOTHING;

-- 10 mata kuliah (TI502 berkuota 2 untuk uji kuota penuh)
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES
    ('TI101', 'Algoritma dan Pemrograman',  3, 1, 30),
    ('TI102', 'Matematika Diskrit',         3, 1, 30),
    ('TI201', 'Struktur Data',              3, 2, 25),
    ('TI202', 'Basis Data',                 4, 2, 25),
    ('TI301', 'Pemrograman Web',            3, 3, 20),
    ('TI302', 'Rekayasa Perangkat Lunak',   3, 3, 20),
    ('TI401', 'Pemrograman Backend Lanjut', 4, 4, 15),
    ('TI402', 'Keamanan Siber',             3, 4, 15),
    ('TI501', 'Kecerdasan Buatan',          3, 5, 10),
    ('TI502', 'Praktikum Proyek Akhir',     2, 5, 2)
ON CONFLICT (kode_mk) DO NOTHING;