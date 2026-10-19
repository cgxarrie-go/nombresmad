package repo

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"nombresmad/backend/internal/model"
)

// AdminRepo stores administrator accounts in Postgres.
type AdminRepo struct {
	db *sql.DB
}

// NewAdminRepo uses the database handle supplied by the caller.
func NewAdminRepo(db *sql.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

// Count returns the number of administrator accounts.
func (r *AdminRepo) Count() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM admins`).Scan(&n)
	return n, err
}

// ByUsername loads one account, including the password hash.
func (r *AdminRepo) ByUsername(username string) (model.Admin, error) {
	var admin model.Admin
	err := r.db.QueryRow(
		`SELECT username, password_hash, must_change_password, credential_version FROM admins WHERE username=$1`,
		username,
	).Scan(&admin.Username, &admin.PasswordHash, &admin.MustChangePassword, &admin.CredentialVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Admin{}, model.ErrNotFound
	}
	return admin, err
}

// List returns every account ordered by username. Password hashes are omitted.
func (r *AdminRepo) List() ([]model.Admin, error) {
	rows, err := r.db.Query(`SELECT username, must_change_password, credential_version FROM admins ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	admins := []model.Admin{}
	for rows.Next() {
		var admin model.Admin
		if err := rows.Scan(&admin.Username, &admin.MustChangePassword, &admin.CredentialVersion); err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return admins, nil
}

// Insert creates an account. The password hash must already be computed.
func (r *AdminRepo) Insert(username, passwordHash string, mustChange bool) error {
	_, err := r.db.Exec(
		`INSERT INTO admins (username, password_hash, must_change_password, credential_version) VALUES ($1, $2, $3, 1)`,
		username, passwordHash, mustChange,
	)
	return mapAdminConstraint(err)
}

// UpdatePassword replaces the hash, sets the change flag, and bumps the session version.
func (r *AdminRepo) UpdatePassword(username, passwordHash string, mustChange bool) (int, error) {
	var version int
	err := r.db.QueryRow(
		`UPDATE admins
		 SET password_hash=$1, must_change_password=$2, credential_version=credential_version+1
		 WHERE username=$3
		 RETURNING credential_version`,
		passwordHash, mustChange, username,
	).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, model.ErrNotFound
	}
	return version, err
}

func mapAdminConstraint(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return model.ErrExists
	}
	return err
}
