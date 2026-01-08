// Package http는 HTTP 연결 및 원격 명령 실행 기능을 제공합니다.
package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	url := fmt.Sprintf("http://%s/respCheck", deviceIP)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
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
	url := fmt.Sprintf("http://%s/agent/req-deploy", deviceIP)

	// 요청 데이터 생성 (템플릿을 변환 없이 그대로 전송)
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
		return nil, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("장비 응답 오류: %d", resp.StatusCode)
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

	// 결과가 하나만 있으면 그것을 반환
	if len(response.Data) == 1 {
		return &response.Data[0], nil
	}

	return nil, fmt.Errorf("장비 %s의 배포 결과를 찾을 수 없습니다", deviceIP)
}

// 장비 상태를 확인합니다. (설정에 따라 Agent 또는 Direct)
func (c *Client) CheckHealth(fw *model.Firewall) (string, error) {
	var isRunning bool
	var err error

	if c.config.IsAgentMode() {
		result, err := c.CheckHealthViaAgent([]string{fw.DeviceName})
		if err != nil {
			return model.ServerStatusStop, err
		}
		isRunning = result[fw.DeviceName]
	} else {
		isRunning, err = c.CheckHealthDirect(fw.DeviceName)
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
	if c.config.IsAgentMode() {
		return c.DeployViaAgent(fw.DeviceName, template)
	}
	return c.DeployDirect(fw.DeviceName, template)
}
