package service

import "testing"

func TestBatasSKS(t *testing.T) {
	cases := []struct {
		name string
		ipk  float64
		want int
	}{
		{"ipk tinggi", 3.45, 24},
		{"tepat 3.00", 3.00, 24},
		{"menengah", 2.75, 21},
		{"tepat 2.50", 2.50, 21},
		{"rendah", 2.49, 18},
		{"nol", 0, 18},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BatasSKS(tc.ipk); got != tc.want {
				t.Errorf("BatasSKS(%v) = %d, diharapkan %d", tc.ipk, got, tc.want)
			}
		})
	}
}

func TestCanAccessStudent(t *testing.T) {
	cases := []struct {
		name          string
		role          string
		authUserID    int
		studentUserID int
		want          bool
	}{
		{"admin boleh semua", "admin", 1, 99, true},
		{"mahasiswa data sendiri", "mahasiswa", 5, 5, true},
		{"mahasiswa data orang lain", "mahasiswa", 5, 6, false},
		{"role tidak dikenal", "tamu", 5, 5, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanAccessStudent(tc.role, tc.authUserID, tc.studentUserID); got != tc.want {
				t.Errorf("CanAccessStudent = %v, diharapkan %v", got, tc.want)
			}
		})
	}
}