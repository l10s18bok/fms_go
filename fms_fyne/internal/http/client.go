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

// 직접 연결로 장비 상태를 확인합니다.
func (c *Client) CheckHealthDirect(fw *model.Firewall) (bool, error) {
	// IP 결정
	ip := fw.DeviceIP
	if ip == "" {
		ip = fw.DeviceName
	}

	// 포트 결정 (장비별 설정 > 전역 설정 > 기본값)
	port := fw.HealthCheckPort
	if port == 0 {
		port = c.config.HealthCheckPort
	}
	if port == 0 {
		port = model.DefaultAPIPort
	}

	// 경로 결정 (장비별 설정 > 전역 설정 > 기본값)
	path := fw.HealthCheckPath
	if path == "" {
		path = c.config.HealthCheckPath
	}
	if path == "" {
		path = model.DefaultHealthCheckPath
	}

	url := fmt.Sprintf("http://%s:%d%s", ip, port, path)
	log.Printf("[DEBUG] CheckHealthDirect: URL=%s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("[ERROR] CheckHealthDirect: 장비 연결 실패 - IP=%s, err=%v", ip, err)
		return false, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	success := resp.StatusCode == http.StatusOK
	log.Printf("[DEBUG] CheckHealthDirect: IP=%s, StatusCode=%d, 성공=%v", ip, resp.StatusCode, success)
	return success, nil
}

// 직접 연결로 방화벽 룰을 배포합니다.
func (c *Client) DeployDirect(fw *model.Firewall, template string) (*model.DeployResult, error) {
	// IP 결정
	ip := fw.DeviceIP
	if ip == "" {
		ip = fw.DeviceName
	}

	// 포트 결정 (장비별 설정 > 전역 설정 > 기본값)
	port := fw.FirewallDeployPort
	if port == 0 {
		port = c.config.FirewallDeployPort
	}
	if port == 0 {
		port = model.DefaultAPIPort
	}

	// 경로 결정 (장비별 설정 > 전역 설정 > 기본값)
	path := fw.FirewallDeployPath
	if path == "" {
		path = c.config.FirewallDeployPath
	}
	if path == "" {
		path = model.DefaultFirewallDeployPath
	}

	url := fmt.Sprintf("http://%s:%d%s", ip, port, path)
	log.Printf("[DEBUG] DeployDirect: URL=%s", url)
	log.Printf("[DEBUG] DeployDirect: template length=%d, content=\n%s", len(template), template)

	// 요청 데이터 생성 (configInfo로 전송)
	reqData := map[string]any{
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
		log.Printf("[ERROR] DeployDirect: 장비 연결 실패 - IP=%s, err=%v", ip, err)
		return nil, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] DeployDirect: 응답 읽기 실패 - %v", err)
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] DeployDirect: 장비 응답 오류 - IP=%s, StatusCode=%d, body=%s", ip, resp.StatusCode, string(body))
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
		if result.IP == ip {
			log.Printf("[DEBUG] DeployDirect: 성공 - IP=%s, Status=%s", ip, result.Status)
			return &result, nil
		}
	}

	// 결과가 하나만 있으면 그것을 반환
	if len(response.Data) == 1 {
		log.Printf("[DEBUG] DeployDirect: 성공 (단일 결과) - IP=%s, Status=%s", ip, response.Data[0].Status)
		return &response.Data[0], nil
	}

	log.Printf("[ERROR] DeployDirect: 장비 결과 없음 - IP=%s", ip)
	return nil, fmt.Errorf("장비 %s의 배포 결과를 찾을 수 없습니다", ip)
}

// 장비 상태를 확인합니다.
func (c *Client) CheckHealth(fw *model.Firewall) (string, error) {
	isRunning, err := c.CheckHealthDirect(fw)
	if err != nil {
		return model.ServerStatusStop, err
	}

	if isRunning {
		return model.ServerStatusRunning, nil
	}
	return model.ServerStatusStop, nil
}

// 방화벽 룰을 배포합니다.
func (c *Client) DeployTemplate(fw *model.Firewall, template string) (*model.DeployResult, error) {
	return c.DeployDirect(fw, template)
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

// GetDeviceInfoDirect 장비 상세 정보를 조회합니다.
func (c *Client) GetDeviceInfoDirect(fw *model.Firewall) (*DeviceReportResponse, error) {
	// IP 결정
	ip := fw.DeviceIP
	if ip == "" {
		ip = fw.DeviceName
	}

	// 포트 결정 (장비별 설정 > 전역 설정 > 기본값)
	port := fw.DeviceInfoPort
	if port == 0 {
		port = c.config.DeviceInfoPort
	}
	if port == 0 {
		port = model.DefaultAPIPort
	}

	// 경로 결정 (장비별 설정 > 전역 설정 > 기본값)
	path := fw.DeviceInfoPath
	if path == "" {
		path = c.config.DeviceInfoPath
	}
	if path == "" {
		path = model.DefaultDeviceInfoPath
	}

	url := fmt.Sprintf("http://%s:%d%s", ip, port, path)
	log.Printf("[DEBUG] GetDeviceInfoDirect: URL=%s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("[ERROR] GetDeviceInfoDirect: 장비 연결 실패 - IP=%s, err=%v", ip, err)
		return nil, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] GetDeviceInfoDirect: 응답 읽기 실패 - %v", err)
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] GetDeviceInfoDirect: 장비 응답 오류 - IP=%s, StatusCode=%d, body=%s", ip, resp.StatusCode, string(body))
		return nil, fmt.Errorf("장비 응답 오류: %d - %s", resp.StatusCode, string(body))
	}

	var report DeviceReportResponse
	if err := json.Unmarshal(body, &report); err != nil {
		log.Printf("[ERROR] GetDeviceInfoDirect: 응답 파싱 실패 - %v, body=%s", err, string(body))
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	log.Printf("[DEBUG] GetDeviceInfoDirect: 성공 - IP=%s, processes=%+v", ip, report.Processes)
	return &report, nil
}

// ProgramUpdateDirect 패키지 업데이트를 요청합니다.
func (c *Client) ProgramUpdateDirect(fw *model.Firewall, req *ProgramUpdateRequest) error {
	// IP 결정
	ip := fw.DeviceIP
	if ip == "" {
		ip = fw.DeviceName
	}

	// 포트 결정 (장비별 설정 > 전역 설정 > 기본값)
	port := fw.ProgramUpdatePort
	if port == 0 {
		port = c.config.ProgramUpdatePort
	}
	if port == 0 {
		port = model.DefaultAPIPort
	}

	// 경로 결정 (장비별 설정 > 전역 설정 > 기본값)
	path := fw.ProgramUpdatePath
	if path == "" {
		path = c.config.ProgramUpdatePath
	}
	if path == "" {
		path = model.DefaultProgramUpdatePath
	}

	url := fmt.Sprintf("http://%s:%d%s", ip, port, path)
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
		log.Printf("[ERROR] ProgramUpdateDirect: 장비 연결 실패 - IP=%s, err=%v", ip, err)
		return fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] ProgramUpdateDirect: 응답 읽기 실패 - %v", err)
		return fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] ProgramUpdateDirect: 장비 응답 오류 - IP=%s, StatusCode=%d, body=%s", ip, resp.StatusCode, string(body))
		return fmt.Errorf("장비 응답 오류: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("[DEBUG] ProgramUpdateDirect: 성공 - IP=%s, response=%s", ip, string(body))
	return nil
}
