package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// JSONRuleFileRepository JSON 기반 규칙 파일 저장소 어댑터
type JSONRuleFileRepository struct {
	store *storage.JSONStore
}

// NewJSONRuleFileRepository 새로운 JSON 규칙 파일 저장소를 생성합니다.
func NewJSONRuleFileRepository(store *storage.JSONStore) RuleFileRepository {
	return &JSONRuleFileRepository{store: store}
}

// GetAll 모든 규칙 파일을 조회합니다.
func (r *JSONRuleFileRepository) GetAll() ([]*model.RuleFile, error) {
	return r.store.GetAllRuleFiles()
}

// GetByVersion 버전으로 규칙 파일을 조회합니다.
func (r *JSONRuleFileRepository) GetByVersion(version string) (*model.RuleFile, error) {
	return r.store.GetRuleFile(version)
}

// Save 규칙 파일을 저장합니다.
func (r *JSONRuleFileRepository) Save(rf *model.RuleFile) error {
	return r.store.SaveRuleFile(rf)
}

// Delete 규칙 파일을 삭제합니다.
func (r *JSONRuleFileRepository) Delete(version string) error {
	return r.store.DeleteRuleFile(version)
}

// Clear 모든 규칙 파일을 삭제합니다.
func (r *JSONRuleFileRepository) Clear() error {
	return r.store.ClearRuleFiles()
}

// Count 규칙 파일 수를 반환합니다.
func (r *JSONRuleFileRepository) Count() int {
	return r.store.CountRuleFiles()
}
