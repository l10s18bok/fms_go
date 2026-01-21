package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// JSONTemplateRepository JSON 기반 템플릿 저장소 어댑터
type JSONTemplateRepository struct {
	store *storage.JSONStore
}

// NewJSONTemplateRepository 새로운 JSON 템플릿 저장소를 생성합니다.
func NewJSONTemplateRepository(store *storage.JSONStore) TemplateRepository {
	return &JSONTemplateRepository{store: store}
}

// GetAll 모든 템플릿을 조회합니다.
func (r *JSONTemplateRepository) GetAll() ([]*model.Template, error) {
	return r.store.GetAllTemplates()
}

// GetByVersion 버전으로 템플릿을 조회합니다.
func (r *JSONTemplateRepository) GetByVersion(version string) (*model.Template, error) {
	return r.store.GetTemplate(version)
}

// Save 템플릿을 저장합니다.
func (r *JSONTemplateRepository) Save(t *model.Template) error {
	return r.store.SaveTemplate(t)
}

// Delete 템플릿을 삭제합니다.
func (r *JSONTemplateRepository) Delete(version string) error {
	return r.store.DeleteTemplate(version)
}

// Clear 모든 템플릿을 삭제합니다.
func (r *JSONTemplateRepository) Clear() error {
	return r.store.ClearTemplates()
}

// Count 템플릿 수를 반환합니다.
func (r *JSONTemplateRepository) Count() int {
	return r.store.CountTemplates()
}
