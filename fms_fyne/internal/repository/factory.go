package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// RepositoryFactory Repository 생성을 담당하는 팩토리
type RepositoryFactory struct {
	config      *model.Config
	jsonStore   *storage.JSONStore
	sqliteStore *storage.SQLiteStore
}

// NewRepositoryFactory 새로운 팩토리를 생성합니다.
func NewRepositoryFactory(config *model.Config, jsonStore *storage.JSONStore, sqliteStore *storage.SQLiteStore) *RepositoryFactory {
	return &RepositoryFactory{
		config:      config,
		jsonStore:   jsonStore,
		sqliteStore: sqliteStore,
	}
}

// NewHistoryRepository 설정에 따라 적절한 이력 저장소를 반환합니다.
func (f *RepositoryFactory) NewHistoryRepository() HistoryRepository {
	if f.config.StorageType == model.StorageTypeSQLite && f.sqliteStore != nil {
		return NewSQLiteHistoryRepository(f.sqliteStore)
	}
	return NewJSONHistoryRepository(f.jsonStore)
}

// IsSQLiteEnabled SQLite 저장소가 활성화되어 있는지 확인합니다.
func (f *RepositoryFactory) IsSQLiteEnabled() bool {
	return f.config.StorageType == model.StorageTypeSQLite && f.sqliteStore != nil
}

// GetSQLiteStore SQLite 저장소를 반환합니다.
func (f *RepositoryFactory) GetSQLiteStore() *storage.SQLiteStore {
	return f.sqliteStore
}

// GetJSONStore JSON 저장소를 반환합니다.
func (f *RepositoryFactory) GetJSONStore() *storage.JSONStore {
	return f.jsonStore
}
