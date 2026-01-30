package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// SQLiteHistoryRepository SQLite 기반 배포 이력 저장소 어댑터
type SQLiteHistoryRepository struct {
	store *storage.SQLiteStore
}

// NewSQLiteHistoryRepository 새로운 SQLite 이력 저장소를 생성합니다.
func NewSQLiteHistoryRepository(store *storage.SQLiteStore) HistoryRepository {
	return &SQLiteHistoryRepository{store: store}
}

// GetAll 모든 이력을 조회합니다.
func (r *SQLiteHistoryRepository) GetAll() ([]*model.DeployHistory, error) {
	return r.store.GetAllHistory()
}

// GetByID ID로 이력을 조회합니다.
func (r *SQLiteHistoryRepository) GetByID(id int) (*model.DeployHistory, error) {
	return r.store.GetHistory(id)
}

// GetByType 이력 유형으로 조회합니다.
func (r *SQLiteHistoryRepository) GetByType(historyType string) ([]*model.DeployHistory, error) {
	return r.store.GetHistoryByType(historyType)
}

// Save 이력을 저장합니다.
func (r *SQLiteHistoryRepository) Save(h *model.DeployHistory) error {
	return r.store.SaveHistory(h)
}

// Delete 이력을 삭제합니다.
func (r *SQLiteHistoryRepository) Delete(id int) error {
	return r.store.DeleteHistory(id)
}

// Clear 모든 이력을 삭제합니다.
func (r *SQLiteHistoryRepository) Clear() error {
	return r.store.ClearHistory()
}

// Count 이력 수를 반환합니다.
func (r *SQLiteHistoryRepository) Count() int {
	return r.store.CountHistory()
}

// GetPage 페이지네이션 기반 이력을 조회합니다.
func (r *SQLiteHistoryRepository) GetPage(req model.PageRequest) (*model.PageResult[model.DeployHistory], error) {
	return r.store.GetHistoryPage(req)
}
