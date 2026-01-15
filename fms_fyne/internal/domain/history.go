package domain

import (
	"strings"
	"time"
)

// JSONTime 커스텀 시간 포맷을 위한 타입 (기존 utils.JSONTime과 호환)
type JSONTime time.Time

const jsonTimeFormat = "2006-01-02 15:04:05"

// MarshalJSON JSONTime을 JSON으로 변환합니다.
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := time.Time(t).Format(jsonTimeFormat)
	return []byte(`"` + stamp + `"`), nil
}

// UnmarshalJSON JSON을 JSONTime으로 변환합니다.
func (t *JSONTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1] // 따옴표 제거

	// 새로운 포맷 시도 (초 포함, 로컬 타임존으로 파싱)
	parsed, err := time.ParseInLocation(jsonTimeFormat, s, time.Local)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	// 기존 포맷 시도 (초 없음, 기존 데이터 호환)
	parsed, err = time.ParseInLocation("2006-01-02 15:04", s, time.Local)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	// 기존 RFC3339 포맷 시도 (기존 데이터 호환)
	parsed, err = time.Parse(time.RFC3339, s)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	return err
}

// Time JSONTime을 time.Time으로 변환합니다.
func (t JSONTime) Time() time.Time {
	return time.Time(t)
}

// Now 현재 시간을 JSONTime으로 반환합니다.
func Now() JSONTime {
	return JSONTime(time.Now())
}

// DeployHistory 배포 이력 엔티티 (확장)
type DeployHistory struct {
	ID         int      `json:"id"`              // 고유 ID (Auto Increment)
	Timestamp  JSONTime `json:"timestamp"`       // 배포 시간
	DeviceName string   `json:"deviceName"`      // 장비명
	DeviceIP   string   `json:"deviceIp"`        // 장비 IP
	Type       string   `json:"type"`            // 이력 유형: "firewall" 또는 "program"
	Version    string   `json:"version"`         // 배포 버전 (룰/프로그램 공통)
	Status     string   `json:"status"`          // 배포 상태 (success/fail/error)

	// 방화벽 룰 배포 전용 (Type="firewall")
	Results []RuleResult `json:"results,omitempty"` // 규칙별 결과

	// 하위 호환성 (기존 필드)
	TemplateVer string `json:"templateVersion,omitempty"` // 기존 템플릿 버전 (읽기 전용)

	// 프로그램 업데이트 전용 (Type="program")
	ProgramName string `json:"programName,omitempty"` // 프로그램 이름
	Message     string `json:"message,omitempty"`     // 상세 메시지
}

// RuleResult 개별 규칙의 배포 결과
type RuleResult struct {
	Rule   string `json:"rule"`   // 규칙 내용
	Text   string `json:"text"`   // 규칙 설명 (optional)
	Status string `json:"status"` // 결과 (ok/error/unfind/validation)
	Reason string `json:"reason"` // 실패 사유
}

// NewDeployHistory 새로운 방화벽 룰 배포 이력을 생성합니다.
func NewDeployHistory(deviceName, deviceIP, version string) *DeployHistory {
	return &DeployHistory{
		ID:         -1,
		Timestamp:  Now(),
		DeviceName: deviceName,
		DeviceIP:   deviceIP,
		Type:       HistoryTypeFirewall,
		Version:    version,
		Status:     DeployStatusUnknown,
		Results:    []RuleResult{},
	}
}

// NewProgramHistory 새로운 프로그램 업데이트 이력을 생성합니다.
func NewProgramHistory(deviceName, deviceIP, programName, version string) *DeployHistory {
	return &DeployHistory{
		ID:          -1,
		Timestamp:   Now(),
		DeviceName:  deviceName,
		DeviceIP:    deviceIP,
		Type:        HistoryTypeProgram,
		Version:     version,
		Status:      DeployStatusUnknown,
		ProgramName: programName,
	}
}

// IsFirewallHistory 방화벽 룰 배포 이력인지 확인합니다.
func (h *DeployHistory) IsFirewallHistory() bool {
	return h.Type == HistoryTypeFirewall || h.Type == ""
}

// IsProgramHistory 프로그램 업데이트 이력인지 확인합니다.
func (h *DeployHistory) IsProgramHistory() bool {
	return h.Type == HistoryTypeProgram
}

// GetVersion 버전을 반환합니다. (하위 호환성 지원)
func (h *DeployHistory) GetVersion() string {
	if h.Version != "" {
		return h.Version
	}
	// 기존 TemplateVer 필드 호환
	return h.TemplateVer
}

// GetType 이력 유형을 반환합니다. (하위 호환성 지원)
func (h *DeployHistory) GetType() string {
	if h.Type != "" {
		return h.Type
	}
	// 기존 이력은 모두 방화벽 룰 배포
	return HistoryTypeFirewall
}

// AddResult 규칙 결과를 추가합니다.
func (h *DeployHistory) AddResult(rule, status, reason string) {
	h.Results = append(h.Results, RuleResult{
		Rule:   rule,
		Status: status,
		Reason: reason,
	})
}

// CalculateStatus 규칙 결과들을 기반으로 전체 상태를 계산합니다.
func (h *DeployHistory) CalculateStatus() {
	if len(h.Results) == 0 {
		h.Status = DeployStatusUnknown
		return
	}

	hasError := false
	for _, r := range h.Results {
		if strings.ToLower(r.Status) != RuleStatusOK {
			hasError = true
			break
		}
	}

	if hasError {
		h.Status = DeployStatusError
	} else {
		h.Status = DeployStatusSuccess
	}
}

// GetTimestampString 포맷된 시간 문자열을 반환합니다.
func (h *DeployHistory) GetTimestampString() string {
	return h.Timestamp.Time().Format(jsonTimeFormat)
}

// GetRuleStatusText 규칙 상태 코드를 표시 텍스트로 변환합니다.
func GetRuleStatusText(status string) string {
	switch strings.ToLower(status) {
	case RuleStatusOK:
		return "OK"
	case RuleStatusError, RuleStatusUnfind, RuleStatusValidation:
		return "실패"
	case RuleStatusWrite:
		return "진행중"
	default:
		return status
	}
}

// GetReasonText 사유를 표시용 텍스트로 변환합니다.
func GetReasonText(reason string) string {
	if reason == "" || strings.ToLower(reason) == "ok" {
		return "-"
	}
	return reason
}

// GetHistoryTypeText 이력 유형을 표시 텍스트로 변환합니다.
func GetHistoryTypeText(historyType string) string {
	switch historyType {
	case HistoryTypeFirewall, "":
		return "방화벽 룰"
	case HistoryTypeProgram:
		return "프로그램"
	default:
		return historyType
	}
}
