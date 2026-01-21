package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// JSONHistoryRepository JSON 기반 배포 이력 저장소 어댑터
type JSONHistoryRepository struct {
	store *storage.JSONStore
}

// NewJSONHistoryRepository 새로운 JSON 이력 저장소를 생성합니다.
func NewJSONHistoryRepository(store *storage.JSONStore) HistoryRepository {
	return &JSONHistoryRepository{store: store}
}

// GetAll 모든 이력을 조회합니다.
func (r *JSONHistoryRepository) GetAll() ([]*model.DeployHistory, error) {
	return r.store.GetAllHistory()
}

// GetByID ID로 이력을 조회합니다.
func (r *JSONHistoryRepository) GetByID(id int) (*model.DeployHistory, error) {
	return r.store.GetHistory(id)
}

// GetByType 이력 유형으로 조회합니다.
func (r *JSONHistoryRepository) GetByType(historyType string) ([]*model.DeployHistory, error) {
	histories, err := r.store.GetAllHistory()
	if err != nil {
		return nil, err
	}

	var filtered []*model.DeployHistory
	for _, h := range histories {
		if h.Type == historyType {
			filtered = append(filtered, h)
		}
	}

	return filtered, nil
}

// Save 이력을 저장합니다.
func (r *JSONHistoryRepository) Save(h *model.DeployHistory) error {
	return r.store.SaveHistory(h)
}

// Delete 이력을 삭제합니다.
func (r *JSONHistoryRepository) Delete(id int) error {
	return r.store.DeleteHistory(id)
}

// Clear 모든 이력을 삭제합니다.
func (r *JSONHistoryRepository) Clear() error {
	return r.store.ClearHistory()
}

// Count 이력 수를 반환합니다.
func (r *JSONHistoryRepository) Count() int {
	return r.store.CountHistory()
}
