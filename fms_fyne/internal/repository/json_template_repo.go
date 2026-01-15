package repository

import (
	"fms/internal/domain"
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
func (r *JSONTemplateRepository) GetAll() ([]*domain.Template, error) {
	templates, err := r.store.GetAllTemplates()
	if err != nil {
		return nil, err
	}
	return r.toDomainList(templates), nil
}

// GetByVersion 버전으로 템플릿을 조회합니다.
func (r *JSONTemplateRepository) GetByVersion(version string) (*domain.Template, error) {
	t, err := r.store.GetTemplate(version)
	if err != nil {
		return nil, err
	}
	return r.toDomain(t), nil
}

// Save 템플릿을 저장합니다.
func (r *JSONTemplateRepository) Save(t *domain.Template) error {
	modelT := r.toModel(t)
	return r.store.SaveTemplate(modelT)
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

// toDomain model.Template을 domain.Template로 변환합니다.
func (r *JSONTemplateRepository) toDomain(t *model.Template) *domain.Template {
	if t == nil {
		return nil
	}
	return &domain.Template{
		Version:  t.Version,
		Contents: t.Contents,
	}
}

// toDomainList model.Template 슬라이스를 domain.Template 슬라이스로 변환합니다.
func (r *JSONTemplateRepository) toDomainList(templates []*model.Template) []*domain.Template {
	result := make([]*domain.Template, len(templates))
	for i, t := range templates {
		result[i] = r.toDomain(t)
	}
	return result
}

// toModel domain.Template을 model.Template로 변환합니다.
func (r *JSONTemplateRepository) toModel(t *domain.Template) *model.Template {
	if t == nil {
		return nil
	}
	return &model.Template{
		Version:  t.Version,
		Contents: t.Contents,
	}
}
