// Package deploy는 방화벽 규칙 배포 기능을 제공합니다.
package deploy

import (
	"sync"

	"fms/internal/http"
	"fms/internal/model"
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

	// 템플릿 전체를 배포
	deployResult, err := d.client.DeployTemplate(fw, template.Contents)
	if err != nil {
		result.Success = false
		result.ErrorMsg = err.Error()
		result.History.Status = model.DeployStatusFail
		fw.DeployStatus = model.DeployStatusFail
		fw.Version = "-"

		// 연결 에러 분석하여 Results에 사유 추가
		errorReason := http.AnalyzeConnectionError(err)
		result.History.Results = append(result.History.Results, model.RuleResult{
			Rule:   "-",
			Text:   "-",
			Status: model.RuleStatusError,
			Reason: errorReason,
		})

		return result
	}

	// DeployResult를 Firewall에 저장
	fw.DeployResult = deployResult

	// DeployResult.Info를 RuleResult로 변환하여 History에 저장
	for _, info := range deployResult.Info {
		ruleResult := model.RuleResult{
			Rule:   info.Rule,
			Text:   info.Text,
			Status: info.Status,
			Reason: info.Reason,
		}
		result.History.Results = append(result.History.Results, ruleResult)
	}

	// 전체 상태 계산
	if deployResult.Status == model.DeployStatusSuccess {
		result.History.Status = model.DeployStatusSuccess
		fw.DeployStatus = model.DeployStatusSuccess
		fw.ServerStatus = model.ServerStatusRunning // 배포 성공 시 서버 상태도 running으로 변경
		fw.Version = template.Version
		result.Success = true
	} else {
		// error 체크
		hasError := false
		for _, info := range deployResult.Info {
			if info.Status != model.RuleStatusOK {
				hasError = true
				break
			}
		}
		if hasError {
			result.History.Status = model.DeployStatusError
			fw.DeployStatus = model.DeployStatusError
		} else {
			result.History.Status = model.DeployStatusFail
			fw.DeployStatus = model.DeployStatusFail
		}
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

// 여러 장비의 연결 상태를 한번에 확인합니다. (Agent 모드용 - 배치 호출, Direct 모드는 병렬 처리)
func (d *Deployer) HealthCheckBatch(firewalls []*model.Firewall) error {
	if len(firewalls) == 0 {
		return nil
	}

	// Agent 모드가 아니면 병렬로 개별 호출 처리
	if !d.config.IsAgentMode() {
		var wg sync.WaitGroup
		for _, fw := range firewalls {
			wg.Add(1)
			go func(f *model.Firewall) {
				defer wg.Done()
				d.HealthCheck(f)
			}(fw)
		}
		wg.Wait()
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
