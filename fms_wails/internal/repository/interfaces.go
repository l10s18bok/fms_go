// Package repository는 데이터 저장소의 인터페이스를 정의합니다.
package repository

import (
	"fms_wails/internal/model"
	"fms_wails/internal/storage"
)

// TemplateRepository는 템플릿 저장소 인터페이스입니다.
type TemplateRepository interface {
	GetAllTemplates() ([]*model.Template, error)
	GetTemplate(version string) (*model.Template, error)
	SaveTemplate(template *model.Template) error
	DeleteTemplate(version string) error
	DeleteAllTemplates() error
}

// FirewallRepository는 장비 저장소 인터페이스입니다.
type FirewallRepository interface {
	GetAllFirewalls() ([]*model.Firewall, error)
	GetFirewall(index int) (*model.Firewall, error)
	SaveFirewall(firewall *model.Firewall) error
	DeleteFirewall(index int) error
	DeleteAllFirewalls() error
}

// HistoryRepository는 배포 이력 저장소 인터페이스입니다.
type HistoryRepository interface {
	GetAllHistory() ([]*model.DeployHistory, error)
	SaveHistory(history *model.DeployHistory) error
	DeleteHistory(id int) error
	DeleteAllHistory() error
}

// ConfigRepository는 설정 저장소 인터페이스입니다.
type ConfigRepository interface {
	GetConfig() (*model.Config, error)
	SaveConfig(config *model.Config) error
}

// DataRepository는 데이터 Import/Export 인터페이스입니다.
type DataRepository interface {
	ExportAll() (*storage.ExportData, error)
	ImportAll(data *storage.ExportData) error
	ClearAll() error
	GetConfigDir() string
}

// Store는 모든 저장소 기능을 통합한 인터페이스입니다.
type Store interface {
	TemplateRepository
	FirewallRepository
	HistoryRepository
	ConfigRepository
	DataRepository
}
