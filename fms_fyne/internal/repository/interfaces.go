// Package repository는 데이터 접근 추상화 계층을 정의합니다.
package repository

import "fms/internal/model"

// FirewallRepository 장비 저장소 인터페이스
type FirewallRepository interface {
	// GetAll 모든 장비를 조회합니다.
	GetAll() ([]*model.Firewall, error)
	// GetByIndex 인덱스로 장비를 조회합니다.
	GetByIndex(index int) (*model.Firewall, error)
	// GetByIP IP 주소로 장비를 조회합니다.
	GetByIP(ip string) (*model.Firewall, error)
	// Save 장비를 저장합니다. (신규 생성 또는 업데이트)
	Save(fw *model.Firewall) error
	// Delete 장비를 삭제합니다.
	Delete(index int) error
	// Clear 모든 장비를 삭제합니다.
	Clear() error
	// Count 장비 수를 반환합니다.
	Count() int
}

// ProgramRepository 패키지 저장소 인터페이스 (신규)
type ProgramRepository interface {
	// GetAll 모든 패키지을 조회합니다.
	GetAll() ([]*model.ProcessInfo, error)
	// GetByID ID로 패키지을 조회합니다.
	GetByID(id int) (*model.ProcessInfo, error)
	// GetByName 이름으로 패키지을 조회합니다.
	GetByName(name string) (*model.ProcessInfo, error)
	// Save 패키지을 저장합니다. (신규 생성 또는 업데이트)
	Save(p *model.ProcessInfo) error
	// Delete 패키지을 삭제합니다.
	Delete(id int) error
	// Clear 모든 패키지을 삭제합니다.
	Clear() error
	// Count 패키지 수를 반환합니다.
	Count() int
}

// HistoryRepository 배포 이력 저장소 인터페이스
type HistoryRepository interface {
	// GetAll 모든 이력을 조회합니다.
	GetAll() ([]*model.DeployHistory, error)
	// GetByID ID로 이력을 조회합니다.
	GetByID(id int) (*model.DeployHistory, error)
	// GetByType 이력 유형으로 조회합니다.
	GetByType(historyType string) ([]*model.DeployHistory, error)
	// Save 이력을 저장합니다.
	Save(h *model.DeployHistory) error
	// Delete 이력을 삭제합니다.
	Delete(id int) error
	// Clear 모든 이력을 삭제합니다.
	Clear() error
	// Count 이력 수를 반환합니다.
	Count() int
}

// TemplateRepository 템플릿 저장소 인터페이스
type TemplateRepository interface {
	// GetAll 모든 템플릿을 조회합니다.
	GetAll() ([]*model.Template, error)
	// GetByVersion 버전으로 템플릿을 조회합니다.
	GetByVersion(version string) (*model.Template, error)
	// Save 템플릿을 저장합니다. (신규 생성 또는 업데이트)
	Save(t *model.Template) error
	// Delete 템플릿을 삭제합니다.
	Delete(version string) error
	// Clear 모든 템플릿을 삭제합니다.
	Clear() error
	// Count 템플릿 수를 반환합니다.
	Count() int
}

// ConfigRepository 설정 저장소 인터페이스
type ConfigRepository interface {
	// Get 설정을 조회합니다.
	Get() (*model.Config, error)
	// Save 설정을 저장합니다.
	Save(c *model.Config) error
}
