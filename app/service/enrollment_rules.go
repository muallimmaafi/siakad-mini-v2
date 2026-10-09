package service

// CanModifyEnrollment: mahasiswa hanya boleh mengubah KRS miliknya sendiri.
// Fungsi murni tanpa Fiber dan repository, jadi mudah diuji.
func CanModifyEnrollment(authUserID, ownerUserID int) bool {
	return authUserID == ownerUserID
}