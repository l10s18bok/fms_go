// Package http는 HTTP 연결 및 원격 명령 실행 기능을 제공합니다.
package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"fms/internal/model"
)

// HTTP 에러를 분석하여 사용자 친화적인 메시지를 반환합니다.
func AnalyzeConnectionError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// 연결 거부 체크 (서버가 명시적으로 거부)
	if strings.Contains(errStr, "connection refused") {
		return "연결 거부"
	}

	// 그 외 모든 경우 (타임아웃, 네트워크 문제, DNS 실패 등)
	return "응답 없음"
}

// Client는 HTTP 클라이언트를 나타냅니다.
type Client struct {
	httpClient *http.Client
	config     *model.Config
}

// 새로운 HTTP 클라이언트를 생성합니다.
func NewClient(config *model.Config) *Client {
	timeout := time.Duration(config.GetTimeoutSeconds()) * time.Second
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		config: config,
	}
}

// Agent 서버를 통해 장비 상태를 확인합니다.
func (c *Client) CheckHealthViaAgent(ipAddrs []string) (map[string]bool, error) {
	url := fmt.Sprintf("%s/agent/req-respCheck", strings.TrimSuffix(c.config.AgentServerURL, "/"))

	// 요청 데이터 생성
	reqData := map[string][]string{
		"ipAddrs": ipAddrs,
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("JSON 변환 실패: %v", err)
	}

	// POST 요청
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Agent 서버 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent 서버 응답 오류: %d", resp.StatusCode)
	}

	// 응답 파싱
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	var result map[string]bool
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	return result, nil
}

// 직접 연결로 장비 상태를 확인합니다.
func (c *Client) CheckHealthDirect(deviceIP string) (bool, error) {
	url := fmt.Sprintf("http://%s/agent/respCheck", deviceIP)
	log.Printf("[DEBUG] CheckHealthDirect: URL=%s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("[ERROR] CheckHealthDirect: 장비 연결 실패 - IP=%s, err=%v", deviceIP, err)
		return false, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	success := resp.StatusCode == http.StatusOK
	log.Printf("[DEBUG] CheckHealthDirect: IP=%s, StatusCode=%d, 성공=%v", deviceIP, resp.StatusCode, success)
	return success, nil
}

// Agent 서버를 통해 템플릿을 배포합니다.
func (c *Client) DeployViaAgent(deviceIP string, template string) (*model.DeployResult, error) {
	url := fmt.Sprintf("%s/agent/req-deploy", strings.TrimSuffix(c.config.AgentServerURL, "/"))

	// 요청 데이터 생성 (index.html과 동일한 형식)
	reqData := map[string]interface{}{
		"template": template,
		"ipAddrs":  []string{deviceIP},
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("JSON 변환 실패: %v", err)
	}

	// POST 요청
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Agent 서버 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent 서버 응답 오류: %d", resp.StatusCode)
	}

	// 응답 파싱
	var response struct {
		Data []model.DeployResult `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	// 해당 장비의 결과 찾기
	for _, result := range response.Data {
		if result.IP == deviceIP {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("장비 %s의 배포 결과를 찾을 수 없습니다", deviceIP)
}

// 직접 연결로 템플릿을 배포합니다.
func (c *Client) DeployDirect(deviceIP string, template string) (*model.DeployResult, error) {
	url := fmt.Sprintf("http://%s/agent/firewall-deploy", deviceIP)
	log.Printf("[DEBUG] DeployDirect: URL=%s", url)
	log.Printf("[DEBUG] DeployDirect: template length=%d, content=\n%s", len(template), template)

	// 요청 데이터 생성 (configInfo로 전송)
	reqData := map[string]interface{}{
		"configInfo": template,
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		log.Printf("[ERROR] DeployDirect: JSON 변환 실패 - %v", err)
		return nil, fmt.Errorf("JSON 변환 실패: %v", err)
	}

	// POST 요청
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ERROR] DeployDirect: 장비 연결 실패 - IP=%s, err=%v", deviceIP, err)
		return nil, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] DeployDirect: 응답 읽기 실패 - %v", err)
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] DeployDirect: 장비 응답 오류 - IP=%s, StatusCode=%d, body=%s", deviceIP, resp.StatusCode, string(body))
		return nil, fmt.Errorf("장비 응답 오류: %d", resp.StatusCode)
	}

	// 응답 파싱
	var response struct {
		Data []model.DeployResult `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("[ERROR] DeployDirect: 응답 파싱 실패 - %v, body=%s", err, string(body))
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	// 해당 장비의 결과 찾기
	for _, result := range response.Data {
		if result.IP == deviceIP {
			log.Printf("[DEBUG] DeployDirect: 성공 - IP=%s, Status=%s", deviceIP, result.Status)
			return &result, nil
		}
	}

	// 결과가 하나만 있으면 그것을 반환
	if len(response.Data) == 1 {
		log.Printf("[DEBUG] DeployDirect: 성공 (단일 결과) - IP=%s, Status=%s", deviceIP, response.Data[0].Status)
		return &response.Data[0], nil
	}

	log.Printf("[ERROR] DeployDirect: 장비 결과 없음 - IP=%s", deviceIP)
	return nil, fmt.Errorf("장비 %s의 배포 결과를 찾을 수 없습니다", deviceIP)
}

// 장비 상태를 확인합니다. (설정에 따라 Agent 또는 Direct)
func (c *Client) CheckHealth(fw *model.Firewall) (string, error) {
	var isRunning bool
	var err error

	// DeviceIP 사용 (DeviceIP가 없으면 DeviceName 사용 - 하위 호환)
	ip := fw.DeviceIP
	if ip == "" {
		ip = fw.DeviceName
	}

	if c.config.IsAgentMode() {
		result, err := c.CheckHealthViaAgent([]string{ip})
		if err != nil {
			return model.ServerStatusStop, err
		}
		isRunning = result[ip]
	} else {
		isRunning, err = c.CheckHealthDirect(ip)
		if err != nil {
			return model.ServerStatusStop, err
		}
	}

	if isRunning {
		return model.ServerStatusRunning, nil
	}
	return model.ServerStatusStop, nil
}

// 템플릿을 배포합니다. (설정에 따라 Agent 또는 Direct)
func (c *Client) DeployTemplate(fw *model.Firewall, template string) (*model.DeployResult, error) {
	// DeviceIP 사용 (DeviceIP가 없으면 DeviceName 사용 - 하위 호환)
	ip := fw.DeviceIP
	if ip == "" {
		ip = fw.DeviceName
	}

	if c.config.IsAgentMode() {
		return c.DeployViaAgent(ip, template)
	}
	return c.DeployDirect(ip, template)
}

// ProgramUpdateRequest 패키지 업데이트 요청 구조체
type ProgramUpdateRequest struct {
	FilePath       string `json:"file_path"`
	ProcessName    string `json:"process_name"`
	ProcessVersion string `json:"process_version"`
}

// DeviceReportResponse 장비 정보 응답 구조체
type DeviceReportResponse struct {
	Processes []ProcessInfo `json:"processes"`
}

// ProcessInfo 프로세스 정보
type ProcessInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// GetDeviceReportDirect 직접 연결로 장비 정보를 조회합니다.
func (c *Client) GetDeviceReportDirect(deviceIP string) (*DeviceReportResponse, error) {
	url := fmt.Sprintf("http://%s/device-report", deviceIP)
	log.Printf("[DEBUG] GetDeviceReportDirect: URL=%s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("[ERROR] GetDeviceReportDirect: 장비 연결 실패 - IP=%s, err=%v", deviceIP, err)
		return nil, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] GetDeviceReportDirect: 응답 읽기 실패 - %v", err)
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] GetDeviceReportDirect: 장비 응답 오류 - IP=%s, StatusCode=%d, body=%s", deviceIP, resp.StatusCode, string(body))
		return nil, fmt.Errorf("장비 응답 오류: %d - %s", resp.StatusCode, string(body))
	}

	var report DeviceReportResponse
	if err := json.Unmarshal(body, &report); err != nil {
		log.Printf("[ERROR] GetDeviceReportDirect: 응답 파싱 실패 - %v, body=%s", err, string(body))
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	log.Printf("[DEBUG] GetDeviceReportDirect: 성공 - IP=%s, processes=%+v", deviceIP, report.Processes)
	return &report, nil
}

// ProgramUpdateDirect 직접 연결로 패키지 업데이트를 요청합니다.
func (c *Client) ProgramUpdateDirect(deviceIP string, req *ProgramUpdateRequest) error {
	url := fmt.Sprintf("http://%s/program-update", deviceIP)
	log.Printf("[DEBUG] ProgramUpdateDirect: URL=%s", url)
	log.Printf("[DEBUG] ProgramUpdateDirect: request=%+v", req)

	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("[ERROR] ProgramUpdateDirect: JSON 변환 실패 - %v", err)
		return fmt.Errorf("JSON 변환 실패: %v", err)
	}
	log.Printf("[DEBUG] ProgramUpdateDirect: JSON=%s", string(jsonData))

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ERROR] ProgramUpdateDirect: 장비 연결 실패 - IP=%s, err=%v", deviceIP, err)
		return fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] ProgramUpdateDirect: 응답 읽기 실패 - %v", err)
		return fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] ProgramUpdateDirect: 장비 응답 오류 - IP=%s, StatusCode=%d, body=%s", deviceIP, resp.StatusCode, string(body))
		return fmt.Errorf("장비 응답 오류: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("[DEBUG] ProgramUpdateDirect: 성공 - IP=%s, response=%s", deviceIP, string(body))
	return nil
}
