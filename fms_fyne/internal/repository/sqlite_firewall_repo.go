package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// SQLiteFirewallRepository SQLite 기반 장비 저장소 어댑터
type SQLiteFirewallRepository struct {
	store *storage.SQLiteStore
}

// NewSQLiteFirewallRepository 새로운 SQLite 장비 저장소를 생성합니다.
func NewSQLiteFirewallRepository(store *storage.SQLiteStore) FirewallRepository {
	return &SQLiteFirewallRepository{store: store}
}

// GetAll 모든 장비를 조회합니다.
func (r *SQLiteFirewallRepository) GetAll() ([]*model.Firewall, error) {
	return r.store.GetAllFirewalls()
}

// GetByIndex 인덱스로 장비를 조회합니다.
func (r *SQLiteFirewallRepository) GetByIndex(index int) (*model.Firewall, error) {
	return r.store.GetFirewall(index)
}

// GetByIP IP 주소로 장비를 조회합니다.
func (r *SQLiteFirewallRepository) GetByIP(ip string) (*model.Firewall, error) {
	return r.store.GetFirewallByIP(ip)
}

// GetPage 페이지네이션 기반 장비를 조회합니다.
func (r *SQLiteFirewallRepository) GetPage(req model.PageRequest) (*model.PageResult[model.Firewall], error) {
	return r.store.GetFirewallPage(req)
}

// Save 장비를 저장합니다.
func (r *SQLiteFirewallRepository) Save(fw *model.Firewall) error {
	return r.store.SaveFirewall(fw)
}

// Delete 장비를 삭제합니다.
func (r *SQLiteFirewallRepository) Delete(index int) error {
	return r.store.DeleteFirewall(index)
}

// Clear 모든 장비를 삭제합니다.
func (r *SQLiteFirewallRepository) Clear() error {
	return r.store.ClearFirewalls()
}

// Count 장비 수를 반환합니다.
func (r *SQLiteFirewallRepository) Count() int {
	return r.store.CountFirewalls()
}
