package service

import (
	"encoding/json"

	"fms_wails/internal/deploy"
	"fms_wails/internal/model"
	"fms_wails/internal/repository"
	"fms_wails/internal/storage"
)

// DeployService는 배포 관련 비즈니스 로직을 처리합니다.
type DeployService struct {
	firewallRepo repository.FirewallRepository
	templateRepo repository.TemplateRepository
	historyRepo  repository.HistoryRepository
	dataRepo     repository.DataRepository
	deployer     *deploy.Deployer
}

// NewDeployService는 새로운 DeployService를 생성합니다.
func NewDeployService(
	firewallRepo repository.FirewallRepository,
	templateRepo repository.TemplateRepository,
	historyRepo repository.HistoryRepository,
	dataRepo repository.DataRepository,
	deployer *deploy.Deployer,
) *DeployService {
	return &DeployService{
		firewallRepo: firewallRepo,
		templateRepo: templateRepo,
		historyRepo:  historyRepo,
		dataRepo:     dataRepo,
		deployer:     deployer,
	}
}

// Deploy는 템플릿을 장비에 배포합니다.
func (s *DeployService) Deploy(firewallIndex int, templateVersion string) (*model.DeployHistory, error) {
	if s.firewallRepo == nil || s.deployer == nil {
		return nil, nil
	}

	firewall, err := s.firewallRepo.GetFirewall(firewallIndex)
	if err != nil {
		return nil, err
	}

	template, err := s.templateRepo.GetTemplate(templateVersion)
	if err != nil {
		return nil, err
	}

	result := s.deployer.Deploy(firewall, template)

	s.historyRepo.SaveHistory(result.History)
	s.firewallRepo.SaveFirewall(firewall)

	return result.History, nil
}

// GetAllHistory는 모든 배포 이력을 반환합니다.
func (s *DeployService) GetAllHistory() []*model.DeployHistory {
	if s.historyRepo == nil {
		return []*model.DeployHistory{}
	}
	history, _ := s.historyRepo.GetAllHistory()
	return history
}

// DeleteHistory는 배포 이력을 삭제합니다.
func (s *DeployService) DeleteHistory(id int) error {
	if s.historyRepo == nil {
		return nil
	}
	return s.historyRepo.DeleteHistory(id)
}

// DeleteAllHistory는 모든 배포 이력을 삭제합니다.
func (s *DeployService) DeleteAllHistory() error {
	if s.historyRepo == nil {
		return nil
	}
	return s.historyRepo.DeleteAllHistory()
}

// SaveHistory는 배포 이력을 저장합니다. (Import용)
func (s *DeployService) SaveHistory(historyJSON string) error {
	if s.historyRepo == nil {
		return nil
	}
	var history model.DeployHistory
	if err := json.Unmarshal([]byte(historyJSON), &history); err != nil {
		return err
	}
	return s.historyRepo.SaveHistory(&history)
}

// ExportData는 모든 데이터를 JSON으로 내보냅니다.
func (s *DeployService) ExportData() string {
	if s.dataRepo == nil {
		return "{}"
	}
	data, _ := s.dataRepo.ExportAll()
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonBytes)
}

// ImportData는 JSON 데이터를 가져옵니다.
func (s *DeployService) ImportData(jsonData string) error {
	if s.dataRepo == nil {
		return nil
	}
	var data storage.ExportData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}
	return s.dataRepo.ImportAll(&data)
}

// ResetAll은 모든 데이터를 초기화합니다.
func (s *DeployService) ResetAll() error {
	if s.dataRepo == nil {
		return nil
	}
	return s.dataRepo.ClearAll()
}

// GetConfigDir은 설정 디렉토리 경로를 반환합니다.
func (s *DeployService) GetConfigDir() string {
	if s.dataRepo == nil {
		return ""
	}
	return s.dataRepo.GetConfigDir()
}
