// Package service는 비즈니스 로직을 처리하는 서비스 계층입니다.
package service

import (
	"fmt"

	"fms_wails/internal/model"
	"fms_wails/internal/repository"
)

// TemplateService는 템플릿 관련 비즈니스 로직을 처리합니다.
type TemplateService struct {
	repo repository.TemplateRepository
}

// NewTemplateService는 새로운 TemplateService를 생성합니다.
func NewTemplateService(repo repository.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

// GetAll은 모든 템플릿을 반환합니다.
func (s *TemplateService) GetAll() []*model.Template {
	if s.repo == nil {
		return []*model.Template{}
	}
	templates, _ := s.repo.GetAllTemplates()
	return templates
}

// Get은 특정 버전의 템플릿을 반환합니다.
func (s *TemplateService) Get(version string) *model.Template {
	if s.repo == nil {
		return nil
	}
	template, err := s.repo.GetTemplate(version)
	if err != nil {
		return nil
	}
	return template
}

// Save는 템플릿을 저장합니다.
func (s *TemplateService) Save(version, contents string) error {
	if s.repo == nil {
		return nil
	}
	template := model.NewTemplate(version, contents)
	if !template.IsValid() {
		return fmt.Errorf("유효하지 않은 템플릿입니다. 버전과 내용을 확인해주세요.")
	}
	return s.repo.SaveTemplate(template)
}

// Delete는 템플릿을 삭제합니다.
func (s *TemplateService) Delete(version string) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.DeleteTemplate(version)
}

// DeleteAll은 모든 템플릿을 삭제합니다.
func (s *TemplateService) DeleteAll() error {
	if s.repo == nil {
		return nil
	}
	return s.repo.DeleteAllTemplates()
}
