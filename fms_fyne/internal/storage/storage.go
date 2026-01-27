// Package storage는 FMS 애플리케이션의 데이터 저장소를 구현합니다.
package storage

import (
	"fms/internal/model"
)

// 데이터 저장소 인터페이스입니다.
type Storage interface {
	// RuleFile 관련 메서드
	GetAllRuleFiles() ([]*model.RuleFile, error)
	GetRuleFile(version string) (*model.RuleFile, error)
	SaveRuleFile(ruleFile *model.RuleFile) error
	DeleteRuleFile(version string) error
	ClearRuleFiles() error

	// Firewall 관련 메서드
	GetAllFirewalls() ([]*model.Firewall, error)
	GetFirewall(index int) (*model.Firewall, error)
	SaveFirewall(firewall *model.Firewall) error
	DeleteFirewall(index int) error
	ClearFirewalls() error

	// DeployHistory 관련 메서드
	GetAllHistory() ([]*model.DeployHistory, error)
	GetHistory(id int) (*model.DeployHistory, error)
	SaveHistory(history *model.DeployHistory) error
	DeleteHistory(id int) error
	ClearHistory() error

	// 전체 데이터 Export/Import
	ExportAll() (*ExportData, error)
	ImportAll(data *ExportData) error
}

// Export/Import용 데이터 구조체입니다.
type ExportData struct {
	RuleFiles []*model.RuleFile      `json:"ruleFiles"`
	Firewalls []*model.Firewall      `json:"firewalls"`
	History   []*model.DeployHistory `json:"history"`
}
