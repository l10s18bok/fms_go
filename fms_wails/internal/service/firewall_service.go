package service

import (
	"encoding/json"

	"fms_wails/internal/deploy"
	"fms_wails/internal/model"
	"fms_wails/internal/repository"
)

// FirewallService는 장비 관련 비즈니스 로직을 처리합니다.
type FirewallService struct {
	repo     repository.FirewallRepository
	deployer *deploy.Deployer
	config   *model.Config
}

// NewFirewallService는 새로운 FirewallService를 생성합니다.
func NewFirewallService(repo repository.FirewallRepository, deployer *deploy.Deployer, config *model.Config) *FirewallService {
	return &FirewallService{
		repo:     repo,
		deployer: deployer,
		config:   config,
	}
}

// GetAll은 모든 장비를 반환합니다.
func (s *FirewallService) GetAll() []*model.Firewall {
	if s.repo == nil {
		return []*model.Firewall{}
	}
	firewalls, _ := s.repo.GetAllFirewalls()
	return firewalls
}

// Get은 특정 장비를 반환합니다.
func (s *FirewallService) Get(index int) *model.Firewall {
	if s.repo == nil {
		return nil
	}
	firewall, err := s.repo.GetFirewall(index)
	if err != nil {
		return nil
	}
	return firewall
}

// Save는 장비를 저장합니다.
func (s *FirewallService) Save(firewallJSON string) error {
	if s.repo == nil {
		return nil
	}
	var firewall model.Firewall
	if err := json.Unmarshal([]byte(firewallJSON), &firewall); err != nil {
		return err
	}
	return s.repo.SaveFirewall(&firewall)
}

// Delete는 장비를 삭제합니다.
func (s *FirewallService) Delete(index int) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.DeleteFirewall(index)
}

// DeleteAll은 모든 장비를 삭제합니다.
func (s *FirewallService) DeleteAll() error {
	if s.repo == nil {
		return nil
	}
	return s.repo.DeleteAllFirewalls()
}

// CheckServerStatus는 단일 장비의 서버 상태를 확인합니다.
func (s *FirewallService) CheckServerStatus(index int) string {
	if s.repo == nil || s.deployer == nil {
		return model.ServerStatusStop
	}

	firewall, err := s.repo.GetFirewall(index)
	if err != nil {
		return model.ServerStatusStop
	}

	s.deployer.HealthCheck(firewall)
	s.repo.SaveFirewall(firewall)

	return firewall.ServerStatus
}

// CheckAllServerStatus는 모든 장비의 상태를 확인합니다.
func (s *FirewallService) CheckAllServerStatus() {
	if s.repo == nil || s.deployer == nil {
		return
	}

	firewalls, _ := s.repo.GetAllFirewalls()
	if len(firewalls) == 0 {
		return
	}

	if s.config.IsAgentMode() {
		s.deployer.HealthCheckBatch(firewalls)
	} else {
		s.deployer.HealthCheckMultiple(firewalls, nil)
	}

	for _, fw := range firewalls {
		s.repo.SaveFirewall(fw)
	}
}

// CheckSelectedServerStatus는 선택된 장비들의 상태를 확인합니다.
func (s *FirewallService) CheckSelectedServerStatus(indexes []int) {
	if s.repo == nil || s.deployer == nil {
		return
	}

	var selectedFirewalls []*model.Firewall
	for _, idx := range indexes {
		fw, err := s.repo.GetFirewall(idx)
		if err == nil && fw != nil {
			selectedFirewalls = append(selectedFirewalls, fw)
		}
	}

	if len(selectedFirewalls) == 0 {
		return
	}

	if s.config.IsAgentMode() {
		s.deployer.HealthCheckBatch(selectedFirewalls)
	} else {
		s.deployer.HealthCheckMultiple(selectedFirewalls, nil)
	}

	for _, fw := range selectedFirewalls {
		s.repo.SaveFirewall(fw)
	}
}

// UpdateConfig는 설정을 업데이트합니다.
func (s *FirewallService) UpdateConfig(config *model.Config) {
	s.config = config
}
