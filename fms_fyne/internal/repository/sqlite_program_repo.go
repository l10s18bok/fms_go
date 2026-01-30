package repository

import (
	"fms/internal/model"
	"fms/internal/storage"
)

// SQLiteProgramRepository SQLite 기반 패키지 저장소 어댑터
type SQLiteProgramRepository struct {
	store *storage.SQLiteStore
}

// NewSQLiteProgramRepository 새로운 SQLite 패키지 저장소를 생성합니다.
func NewSQLiteProgramRepository(store *storage.SQLiteStore) ProgramRepository {
	return &SQLiteProgramRepository{store: store}
}

// GetAll 모든 패키지를 조회합니다.
func (r *SQLiteProgramRepository) GetAll() ([]*model.ProcessInfo, error) {
	return r.store.GetAllPrograms()
}

// GetByID ID로 패키지를 조회합니다.
func (r *SQLiteProgramRepository) GetByID(id int) (*model.ProcessInfo, error) {
	return r.store.GetProgram(id)
}

// GetByName 이름으로 패키지를 조회합니다.
func (r *SQLiteProgramRepository) GetByName(name string) (*model.ProcessInfo, error) {
	return r.store.GetProgramByName(name)
}

// GetPage 페이지네이션 기반 패키지를 조회합니다.
func (r *SQLiteProgramRepository) GetPage(req model.PageRequest) (*model.PageResult[model.ProcessInfo], error) {
	return r.store.GetProgramPage(req)
}

// Save 패키지를 저장합니다.
func (r *SQLiteProgramRepository) Save(p *model.ProcessInfo) error {
	return r.store.SaveProgram(p)
}

// Delete 패키지를 삭제합니다.
func (r *SQLiteProgramRepository) Delete(id int) error {
	return r.store.DeleteProgram(id)
}

// Clear 모든 패키지를 삭제합니다.
func (r *SQLiteProgramRepository) Clear() error {
	return r.store.ClearPrograms()
}

// Count 패키지 수를 반환합니다.
func (r *SQLiteProgramRepository) Count() int {
	return r.store.CountPrograms()
}
