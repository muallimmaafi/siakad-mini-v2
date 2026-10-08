package helper

import "golang.org/x/crypto/bcrypt"

// bcryptCost menentukan berapa kali proses hashing diulang secara internal.
// Semakin besar, semakin lambat pula percobaan menebak password satu per satu.
const bcryptCost = 10

// dummyHash dipakai ketika email tidak ditemukan, agar waktu tanggap login
// tetap mirip dengan kasus password salah. Tanpa ini, selisih waktu respons
// membocorkan email mana yang terdaftar (timing attack).
var dummyHash = mustDummyHash()

func mustDummyHash() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcryptCost)
	if err != nil {
		panic("gagal membuat dummy hash: " + err.Error())
	}
	return hash
}

// HashPassword mengubah password menjadi hash yang tidak dapat dikembalikan.
// bcrypt menyisipkan salt acak, sehingga dua user dengan password sama
// tetap menghasilkan hash yang berbeda.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword membandingkan password dengan hash-nya.
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// VerifyDummyPassword sengaja membuang waktu seperti VerifyPassword.
func VerifyDummyPassword(plain string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plain))
}