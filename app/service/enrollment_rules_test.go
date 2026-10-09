package service

import "testing"

func TestIsTahunAkademik(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"2026/2027-Ganjil", true},
		{"2026/2027-Genap", true},
		{"2026/2028-Ganjil", false}, // tahun kedua harus +1
		{"2026-Ganjil", false},
		{"2026/2027-ganjil", false},
		{"2026/2027", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			if got := IsTahunAkademik(tc.input); got != tc.want {
				t.Errorf("IsTahunAkademik(%q) = %v, diharapkan %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestCanModifyEnrollment(t *testing.T) {
	if !CanModifyEnrollment(5, 5) {
		t.Error("pemilik seharusnya boleh mengubah KRS-nya")
	}
	if CanModifyEnrollment(5, 6) {
		t.Error("mahasiswa lain seharusnya tidak boleh mengubah KRS orang lain")
	}
}