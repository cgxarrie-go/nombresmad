package service

import (
	"sort"
	"sync"

	"nombresmad/backend/internal/model"
)

// AdminStore is the persistence used by administrator accounts.
type AdminStore interface {
	Count() (int, error)
	ByUsername(username string) (model.Admin, error)
	List() ([]model.Admin, error)
	Insert(username, passwordHash string, mustChange bool) error
	UpdatePassword(username, passwordHash string, mustChange bool) (int, error)
}

type memoryAdmins struct {
	mu   sync.Mutex
	rows []model.Admin
}

// NewMemoryAdminStore returns an in-memory store for tests.
func NewMemoryAdminStore() AdminStore {
	return &memoryAdmins{}
}

func (m *memoryAdmins) Count() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.rows), nil
}

func (m *memoryAdmins) ByUsername(username string) (model.Admin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.byUsername(username)
}

func (m *memoryAdmins) byUsername(username string) (model.Admin, error) {
	for _, row := range m.rows {
		if row.Username == username {
			return row, nil
		}
	}
	return model.Admin{}, model.ErrNotFound
}

func (m *memoryAdmins) List() ([]model.Admin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Admin, len(m.rows))
	copy(out, m.rows)
	sort.Slice(out, func(i, j int) bool { return out[i].Username < out[j].Username })
	return out, nil
}

func (m *memoryAdmins) Insert(username, passwordHash string, mustChange bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := m.byUsername(username); err == nil {
		return model.ErrExists
	}
	m.rows = append(m.rows, model.Admin{
		Username:           username,
		PasswordHash:       passwordHash,
		MustChangePassword: mustChange,
		CredentialVersion:  1,
	})
	return nil
}

func (m *memoryAdmins) UpdatePassword(username, passwordHash string, mustChange bool) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.rows {
		if m.rows[i].Username == username {
			m.rows[i].PasswordHash = passwordHash
			m.rows[i].MustChangePassword = mustChange
			m.rows[i].CredentialVersion++
			return m.rows[i].CredentialVersion, nil
		}
	}
	return 0, model.ErrNotFound
}
