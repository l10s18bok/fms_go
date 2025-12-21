// Package deploy는 방화벽 규칙 배포 기능을 제공합니다.
package deploy

import (
	"strings"
	"sync"

	"fms_wails/internal/http"
	"fms_wails/internal/model"
)

// 배포를 관리합니다.
type Deployer struct {
	mu     sync.Mutex
	config *model.Config
	client *http.Client
}

// 새로운 Deployer를 생성합니다.
func NewDeployer(config *model.Config) *Deployer {
	return &Deployer{
		config: config,
		client: http.NewClient(config),
	}
}

// 단일 장비의 배포 결과를 나타냅니다.
type DeployResult struct {
	Firewall *model.Firewall
	History  *model.DeployHistory
	Success  bool
	ErrorMsg string
}

// 단일 장비에 템플릿을 배포합니다.
func (d *Deployer) Deploy(fw *model.Firewall, template *model.Template) *DeployResult {
	d.mu.Lock()
	defer d.mu.Unlock()

	result := &DeployResult{
		Firewall: fw,
		History:  model.NewDeployHistory(fw.DeviceName, template.Version),
	}

	// 템플릿 내용을 규칙 단위로 분리
	rules := strings.Split(template.Contents, "\n")

	// 규칙 배포
	ruleResults, err := d.client.DeployRules(fw, rules)
	if err != nil {
		result.Success = false
		result.ErrorMsg = err.Error()
		result.History.Status = model.DeployStatusFail
		fw.DeployStatus = model.DeployStatusFail
		fw.Version = "-"
		return result
	}

	// 결과 저장
	result.History.Results = ruleResults

	// 전체 상태 계산
	result.History.CalculateStatus()

	// 장비 상태 업데이트
	fw.DeployStatus = result.History.Status
	if result.History.Status == model.DeployStatusSuccess {
		fw.Version = template.Version
		result.Success = true
	} else {
		fw.Version = "-"
		result.Success = false
	}

	return result
}

// 여러 장비에 템플릿을 배포합니다.
func (d *Deployer) DeployToMultiple(firewalls []*model.Firewall, template *model.Template, progressCb func(int, int, string)) []*DeployResult {
	results := make([]*DeployResult, 0, len(firewalls))
	total := len(firewalls)

	for i, fw := range firewalls {
		if progressCb != nil {
			progressCb(i+1, total, fw.DeviceName)
		}

		result := d.Deploy(fw, template)
		results = append(results, result)
	}

	return results
}

// 장비의 연결 상태를 확인합니다.
func (d *Deployer) HealthCheck(fw *model.Firewall) error {
	status, err := d.client.CheckHealth(fw)
	fw.ServerStatus = status
	return err
}

// 여러 장비의 연결 상태를 확인합니다. (Direct 모드용 - 개별 호출)
func (d *Deployer) HealthCheckMultiple(firewalls []*model.Firewall, progressCb func(int, int, string)) map[int]error {
	errors := make(map[int]error)
	total := len(firewalls)

	for i, fw := range firewalls {
		if progressCb != nil {
			progressCb(i+1, total, fw.DeviceName)
		}

		err := d.HealthCheck(fw)
		if err != nil {
			errors[fw.Index] = err
		}
	}

	return errors
}

// 여러 장비의 연결 상태를 한번에 확인합니다. (Agent 모드용 - 배치 호출)
func (d *Deployer) HealthCheckBatch(firewalls []*model.Firewall) error {
	if len(firewalls) == 0 {
		return nil
	}

	// Agent 모드가 아니면 개별 호출로 처리
	if !d.config.IsAgentMode() {
		for _, fw := range firewalls {
			d.HealthCheck(fw)
		}
		return nil
	}

	// 모든 장비의 IP 주소 수집
	ipAddrs := make([]string, len(firewalls))
	for i, fw := range firewalls {
		ipAddrs[i] = fw.DeviceName
	}

	// Agent 서버에 한번에 요청
	results, err := d.client.CheckHealthViaAgent(ipAddrs)
	if err != nil {
		// 에러 시 모든 장비를 stop 상태로 설정
		for _, fw := range firewalls {
			fw.ServerStatus = model.ServerStatusStop
		}
		return err
	}

	// 결과를 각 장비에 적용
	for _, fw := range firewalls {
		if isRunning, ok := results[fw.DeviceName]; ok && isRunning {
			fw.ServerStatus = model.ServerStatusRunning
		} else {
			fw.ServerStatus = model.ServerStatusStop
		}
	}

	return nil
}

// 설정을 업데이트합니다.
func (d *Deployer) UpdateConfig(config *model.Config) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config = config
	d.client = http.NewClient(config)
}
