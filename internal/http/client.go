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
	log.Printf("[DEBUG] CheckHealthDirect 요청: %s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("[DEBUG] CheckHealthDirect 연결 실패: %v", err)
		return false, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[DEBUG] CheckHealthDirect 응답: StatusCode=%d, Body=%s", resp.StatusCode, string(body))

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

// 템플릿을 Direct 모드용 형식으로 변환합니다.
// 입력: req|INSERT|101|INPUT|ACCEPT|192.168.44.11|TCP|ANY|9090||
// 출력: agent -m=insert -c=INPUT -p=tcp --dport=9090 -a=ACCEPT -s=192.168.44.11
func convertTemplateForDirect(template string) string {
	lines := strings.Split(template, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// req|INSERT|101|INPUT|ACCEPT|192.168.44.11|TCP|ANY|9090||
		parts := strings.Split(line, "|")
		if len(parts) < 9 {
			continue
		}

		// parts[0]: req
		// parts[1]: INSERT/DELETE/FLUSH 등
		// parts[2]: ID
		// parts[3]: CHAIN (INPUT/OUTPUT/FORWARD)
		// parts[4]: ACTION (ACCEPT/DROP/REJECT)
		// parts[5]: SRC IP
		// parts[6]: PROTOCOL (TCP/UDP/ANY)
		// parts[7]: DST IP (또는 ANY)
		// parts[8]: DPORT

		method := strings.ToLower(parts[1])
		chain := parts[3]
		action := parts[4]
		srcIP := parts[5]
		protocol := strings.ToLower(parts[6])
		dport := parts[8]

		// agent 명령어 형식으로 변환
		cmd := fmt.Sprintf("agent -m=%s -c=%s", method, chain)

		if protocol != "any" && protocol != "" {
			cmd += fmt.Sprintf(" -p=%s", protocol)
		}

		if dport != "" && dport != "ANY" {
			cmd += fmt.Sprintf(" --dport=%s", dport)
		}

		cmd += fmt.Sprintf(" -a=%s", action)

		if srcIP != "" && srcIP != "ANY" {
			cmd += fmt.Sprintf(" -s=%s", srcIP)
		}

		result = append(result, cmd)
	}

	return strings.Join(result, "\n")
}

// 직접 연결로 템플릿을 배포합니다.
func (c *Client) DeployDirect(deviceIP string, template string) (*model.DeployResult, error) {
	url := fmt.Sprintf("http://%s/agent/req-deploy", deviceIP)
	log.Printf("[DEBUG] DeployDirect 요청: %s", url)

	// Direct 모드용 템플릿 형식으로 변환
	convertedTemplate := convertTemplateForDirect(template)
	log.Printf("[DEBUG] DeployDirect 변환된 템플릿:\n%s", convertedTemplate)

	// 요청 데이터 생성
	reqData := map[string]interface{}{
		"template": convertedTemplate,
		"ipAddrs":  []string{deviceIP},
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("JSON 변환 실패: %v", err)
	}
	log.Printf("[DEBUG] DeployDirect 요청 Body: %s", string(jsonData))

	// POST 요청
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[DEBUG] DeployDirect 연결 실패: %v", err)
		return nil, fmt.Errorf("장비 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}
	log.Printf("[DEBUG] DeployDirect 응답: StatusCode=%d, Body=%s", resp.StatusCode, string(body))

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
	log.Printf("[DEBUG] DeployDirect 파싱 결과: Data 개수=%d", len(response.Data))

	// 해당 장비의 결과 찾기
	for _, result := range response.Data {
		log.Printf("[DEBUG] DeployDirect 결과 IP=%s, Status=%s", result.IP, result.Status)
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
