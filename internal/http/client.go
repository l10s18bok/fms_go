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

// Agent 서버를 통해 규칙을 배포합니다.
func (c *Client) DeployViaAgent(deviceIP string, rules []string) ([]model.RuleResult, error) {
	url := fmt.Sprintf("%s/agent/req-deploy", strings.TrimSuffix(c.config.AgentServerURL, "/"))

	results := make([]model.RuleResult, 0, len(rules))

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		result := model.RuleResult{
			Rule:   rule,
			Status: model.RuleStatusOK,
		}

		// 요청 데이터 생성
		reqData := map[string]string{
			"deviceName": deviceIP,
			"rule":       rule,
		}
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			result.Status = model.RuleStatusError
			result.Reason = fmt.Sprintf("JSON 변환 실패: %v", err)
			results = append(results, result)
			continue
		}

		// POST 요청
		resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			result.Status = model.RuleStatusError
			result.Reason = fmt.Sprintf("Agent 서버 연결 실패: %v", err)
			results = append(results, result)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			result.Status = model.RuleStatusError
			result.Reason = fmt.Sprintf("Agent 서버 응답 오류: %d", resp.StatusCode)
			results = append(results, result)
			continue
		}

		// 응답 확인 (OK 포함 여부)
		outputStr := strings.ToUpper(string(body))
		if strings.Contains(outputStr, "OK") {
			result.Status = model.RuleStatusOK
			result.Reason = ""
		} else {
			result.Status = model.RuleStatusError
			result.Reason = fmt.Sprintf("응답: %s", strings.TrimSpace(string(body)))
		}

		results = append(results, result)
	}

	return results, nil
}

// 직접 연결로 규칙을 배포합니다.
func (c *Client) DeployDirect(deviceIP string, rules []string) ([]model.RuleResult, error) {
	url := fmt.Sprintf("http://%s/deploy", deviceIP)

	results := make([]model.RuleResult, 0, len(rules))

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		result := model.RuleResult{
			Rule:   rule,
			Status: model.RuleStatusOK,
		}

		// 요청 데이터 생성
		reqData := map[string]string{
			"rule": rule,
		}
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			result.Status = model.RuleStatusError
			result.Reason = fmt.Sprintf("JSON 변환 실패: %v", err)
			results = append(results, result)
			continue
		}

		// POST 요청
		resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			result.Status = model.RuleStatusError
			result.Reason = fmt.Sprintf("장비 연결 실패: %v", err)
			results = append(results, result)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			result.Status = model.RuleStatusError
			result.Reason = fmt.Sprintf("장비 응답 오류: %d", resp.StatusCode)
			results = append(results, result)
			continue
		}

		// 응답 확인 (OK 포함 여부)
		outputStr := strings.ToUpper(string(body))
		if strings.Contains(outputStr, "OK") {
			result.Status = model.RuleStatusOK
			result.Reason = ""
		} else {
			result.Status = model.RuleStatusError
			result.Reason = fmt.Sprintf("응답: %s", strings.TrimSpace(string(body)))
		}

		results = append(results, result)
	}

	return results, nil
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

// 규칙을 배포합니다. (설정에 따라 Agent 또는 Direct)
func (c *Client) DeployRules(fw *model.Firewall, rules []string) ([]model.RuleResult, error) {
	if c.config.IsAgentMode() {
		return c.DeployViaAgent(fw.DeviceName, rules)
	}
	return c.DeployDirect(fw.DeviceName, rules)
}
