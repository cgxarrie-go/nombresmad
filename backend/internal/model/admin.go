package model

// Admin is a stored administrator. PasswordHash is a one-way password hash.
type Admin struct {
	Username           string
	PasswordHash       string
	MustChangePassword bool
	CredentialVersion  int
}

// AdminView is an administrator without credential material.
type AdminView struct {
	Username           string `json:"username"`
	MustChangePassword bool   `json:"mustChangePassword"`
}
