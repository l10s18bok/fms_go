package repository

import (
	"fms/internal/domain"
	"fms/internal/model"
	"fms/internal/storage"
)

// JSONProgramRepository JSON 기반 프로그램 저장소 어댑터
type JSONProgramRepository struct {
	store *storage.JSONStore
}

// NewJSONProgramRepository 새로운 JSON 프로그램 저장소를 생성합니다.
func NewJSONProgramRepository(store *storage.JSONStore) ProgramRepository {
	return &JSONProgramRepository{store: store}
}

// GetAll 모든 프로그램을 조회합니다.
func (r *JSONProgramRepository) GetAll() ([]*domain.ProcessInfo, error) {
	programs, err := r.store.GetAllPrograms()
	if err != nil {
		return nil, err
	}
	return r.toDomainList(programs), nil
}

// GetByID ID로 프로그램을 조회합니다.
func (r *JSONProgramRepository) GetByID(id int) (*domain.ProcessInfo, error) {
	p, err := r.store.GetProgram(id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(p), nil
}

// GetByName 이름으로 프로그램을 조회합니다.
func (r *JSONProgramRepository) GetByName(name string) (*domain.ProcessInfo, error) {
	p, err := r.store.GetProgramByName(name)
	if err != nil {
		return nil, err
	}
	return r.toDomain(p), nil
}

// Save 프로그램을 저장합니다.
func (r *JSONProgramRepository) Save(p *domain.ProcessInfo) error {
	modelP := r.toModel(p)
	err := r.store.SaveProgram(modelP)
	if err != nil {
		return err
	}
	// ID 업데이트 (새로 생성된 경우)
	p.ID = modelP.ID
	return nil
}

// Delete 프로그램을 삭제합니다.
func (r *JSONProgramRepository) Delete(id int) error {
	return r.store.DeleteProgram(id)
}

// Clear 모든 프로그램을 삭제합니다.
func (r *JSONProgramRepository) Clear() error {
	return r.store.ClearPrograms()
}

// Count 프로그램 수를 반환합니다.
func (r *JSONProgramRepository) Count() int {
	return r.store.CountPrograms()
}

// toDomain model.ProcessInfo를 domain.ProcessInfo로 변환합니다.
func (r *JSONProgramRepository) toDomain(p *model.ProcessInfo) *domain.ProcessInfo {
	if p == nil {
		return nil
	}
	return &domain.ProcessInfo{
		ID:                p.ID,
		ProcessName:       p.ProcessName,
		ProcessFilePath:   p.ProcessFilePath,
		ProcessUploadPath: p.ProcessUploadPath,
		ProcessVersion:    p.ProcessVersion,
		ProcessCreatedAt:  p.ProcessCreatedAt,
	}
}

// toDomainList model.ProcessInfo 슬라이스를 domain.ProcessInfo 슬라이스로 변환합니다.
func (r *JSONProgramRepository) toDomainList(programs []*model.ProcessInfo) []*domain.ProcessInfo {
	result := make([]*domain.ProcessInfo, len(programs))
	for i, p := range programs {
		result[i] = r.toDomain(p)
	}
	return result
}

// toModel domain.ProcessInfo를 model.ProcessInfo로 변환합니다.
func (r *JSONProgramRepository) toModel(p *domain.ProcessInfo) *model.ProcessInfo {
	if p == nil {
		return nil
	}
	return &model.ProcessInfo{
		ID:                p.ID,
		ProcessName:       p.ProcessName,
		ProcessFilePath:   p.ProcessFilePath,
		ProcessUploadPath: p.ProcessUploadPath,
		ProcessVersion:    p.ProcessVersion,
		ProcessCreatedAt:  p.ProcessCreatedAt,
	}
}
