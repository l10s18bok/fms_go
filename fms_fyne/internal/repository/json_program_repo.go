package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// JSONProgramRepository JSON 기반 패키지 저장소 어댑터
type JSONProgramRepository struct {
	store *storage.JSONStore
}

// NewJSONProgramRepository 새로운 JSON 패키지 저장소를 생성합니다.
func NewJSONProgramRepository(store *storage.JSONStore) ProgramRepository {
	return &JSONProgramRepository{store: store}
}

// GetAll 모든 패키지을 조회합니다.
func (r *JSONProgramRepository) GetAll() ([]*model.ProcessInfo, error) {
	return r.store.GetAllPrograms()
}

// GetByID ID로 패키지을 조회합니다.
func (r *JSONProgramRepository) GetByID(id int) (*model.ProcessInfo, error) {
	return r.store.GetProgram(id)
}

// GetByName 이름으로 패키지을 조회합니다.
func (r *JSONProgramRepository) GetByName(name string) (*model.ProcessInfo, error) {
	return r.store.GetProgramByName(name)
}

// Save 패키지을 저장합니다.
func (r *JSONProgramRepository) Save(p *model.ProcessInfo) error {
	return r.store.SaveProgram(p)
}

// Delete 패키지을 삭제합니다.
func (r *JSONProgramRepository) Delete(id int) error {
	return r.store.DeleteProgram(id)
}

// Clear 모든 패키지을 삭제합니다.
func (r *JSONProgramRepository) Clear() error {
	return r.store.ClearPrograms()
}

// Count 패키지 수를 반환합니다.
func (r *JSONProgramRepository) Count() int {
	return r.store.CountPrograms()
}
