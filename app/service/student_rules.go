package service

// BatasSKS menentukan batas maksimal SKS per semester dari IPK terakhir.
func BatasSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	}
	if ipk >= 2.50 {
		return 21
	}
	return 18
}

// CanAccessStudent: admin boleh mengakses semua data mahasiswa,
// mahasiswa hanya boleh mengakses data miliknya sendiri.
func CanAccessStudent(role string, authUserID, studentUserID int) bool {
	if role == "admin" {
		return true
	}
	return role == "mahasiswa" && authUserID == studentUserID
}