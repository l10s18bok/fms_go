// Package repository는 데이터 접근 추상화 계층을 정의합니다.
package repository

import "fms/internal/domain"

// FirewallRepository 장비 저장소 인터페이스
type FirewallRepository interface {
	// GetAll 모든 장비를 조회합니다.
	GetAll() ([]*domain.Firewall, error)
	// GetByIndex 인덱스로 장비를 조회합니다.
	GetByIndex(index int) (*domain.Firewall, error)
	// GetByIP IP 주소로 장비를 조회합니다.
	GetByIP(ip string) (*domain.Firewall, error)
	// Save 장비를 저장합니다. (신규 생성 또는 업데이트)
	Save(fw *domain.Firewall) error
	// Delete 장비를 삭제합니다.
	Delete(index int) error
	// Clear 모든 장비를 삭제합니다.
	Clear() error
	// Count 장비 수를 반환합니다.
	Count() int
}

// ProgramRepository 프로그램 저장소 인터페이스 (신규)
type ProgramRepository interface {
	// GetAll 모든 프로그램을 조회합니다.
	GetAll() ([]*domain.ProcessInfo, error)
	// GetByID ID로 프로그램을 조회합니다.
	GetByID(id int) (*domain.ProcessInfo, error)
	// GetByName 이름으로 프로그램을 조회합니다.
	GetByName(name string) (*domain.ProcessInfo, error)
	// Save 프로그램을 저장합니다. (신규 생성 또는 업데이트)
	Save(p *domain.ProcessInfo) error
	// Delete 프로그램을 삭제합니다.
	Delete(id int) error
	// Clear 모든 프로그램을 삭제합니다.
	Clear() error
	// Count 프로그램 수를 반환합니다.
	Count() int
}

// HistoryRepository 배포 이력 저장소 인터페이스
type HistoryRepository interface {
	// GetAll 모든 이력을 조회합니다.
	GetAll() ([]*domain.DeployHistory, error)
	// GetByID ID로 이력을 조회합니다.
	GetByID(id int) (*domain.DeployHistory, error)
	// GetByType 이력 유형으로 조회합니다.
	GetByType(historyType string) ([]*domain.DeployHistory, error)
	// Save 이력을 저장합니다.
	Save(h *domain.DeployHistory) error
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
	GetAll() ([]*domain.Template, error)
	// GetByVersion 버전으로 템플릿을 조회합니다.
	GetByVersion(version string) (*domain.Template, error)
	// Save 템플릿을 저장합니다. (신규 생성 또는 업데이트)
	Save(t *domain.Template) error
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
	Get() (*domain.Config, error)
	// Save 설정을 저장합니다.
	Save(c *domain.Config) error
}
