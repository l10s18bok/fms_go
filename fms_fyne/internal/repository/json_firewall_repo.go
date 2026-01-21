package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// JSONFirewallRepository JSON 기반 장비 저장소 어댑터
type JSONFirewallRepository struct {
	store *storage.JSONStore
}

// NewJSONFirewallRepository 새로운 JSON 장비 저장소를 생성합니다.
func NewJSONFirewallRepository(store *storage.JSONStore) FirewallRepository {
	return &JSONFirewallRepository{store: store}
}

// GetAll 모든 장비를 조회합니다.
func (r *JSONFirewallRepository) GetAll() ([]*model.Firewall, error) {
	return r.store.GetAllFirewalls()
}

// GetByIndex 인덱스로 장비를 조회합니다.
func (r *JSONFirewallRepository) GetByIndex(index int) (*model.Firewall, error) {
	return r.store.GetFirewall(index)
}

// GetByIP IP 주소로 장비를 조회합니다.
func (r *JSONFirewallRepository) GetByIP(ip string) (*model.Firewall, error) {
	return r.store.GetFirewallByIP(ip)
}

// Save 장비를 저장합니다.
func (r *JSONFirewallRepository) Save(fw *model.Firewall) error {
	return r.store.SaveFirewall(fw)
}

// Delete 장비를 삭제합니다.
func (r *JSONFirewallRepository) Delete(index int) error {
	return r.store.DeleteFirewall(index)
}

// Clear 모든 장비를 삭제합니다.
func (r *JSONFirewallRepository) Clear() error {
	return r.store.ClearFirewalls()
}

// Count 장비 수를 반환합니다.
func (r *JSONFirewallRepository) Count() int {
	return r.store.CountFirewalls()
}
