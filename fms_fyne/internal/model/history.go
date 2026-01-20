package model

import (
	"strings"

	"fms/internal/utils"
)

// 이력 유형 상수
const (
	HistoryTypeFirewall = "firewall" // 방화벽 룰 배포
	HistoryTypeProgram  = "program"  // 패키지 업데이트
)

// 배포 이력을 나타냅니다. (PRD 3.3.4 기준)
type DeployHistory struct {
	ID          int            `json:"id"`                       // 고유 ID (Auto Increment)
	Type        string         `json:"type,omitempty"`           // 이력 유형 (firewall/program)
	Timestamp   utils.JSONTime `json:"timestamp"`                // 배포 시간
	DeviceName  string         `json:"deviceName,omitempty"`     // 장비명 (PRD 3.3.4)
	DeviceIP    string         `json:"deviceIp"`                 // 장비 IP
	TemplateVer string         `json:"templateVersion"`          // 배포한 템플릿 버전 (방화벽 룰)
	ProgramName string         `json:"programName,omitempty"`    // 패키지 이름 (패키지 업데이트)
	ProgramVer  string         `json:"programVersion,omitempty"` // 패키지 버전 (패키지 업데이트)
	Message     string         `json:"message,omitempty"`        // 결과 메시지
	Status      string         `json:"status"`                   // 배포 상태 (success/fail/error)
	Results     []RuleResult   `json:"results"`                  // 규칙별 결과 (방화벽 룰)
}

// 개별 규칙의 배포 결과를 나타냅니다.
type RuleResult struct {
	Rule   string `json:"rule"`   // 규칙 내용
	Text   string `json:"text"`   // 규칙 설명 (optional)
	Status string `json:"status"` // 결과 (ok/error/unfind/validation)
	Reason string `json:"reason"` // 실패 사유
}

// 규칙 결과 상태 상수
const (
	RuleStatusOK         = "ok"
	RuleStatusError      = "error"
	RuleStatusUnfind     = "unfind"
	RuleStatusValidation = "validation"
	RuleStatusWrite      = "write"
)

// 새로운 방화벽 룰 배포 이력을 생성합니다.
func NewDeployHistory(deviceName, deviceIP, templateVer string) *DeployHistory {
	return &DeployHistory{
		Type:        HistoryTypeFirewall,
		Timestamp:   utils.Now(),
		DeviceName:  deviceName,
		DeviceIP:    deviceIP,
		TemplateVer: templateVer,
		Status:      DeployStatusUnknown,
		Results:     []RuleResult{},
	}
}

// 새로운 패키지 업데이트 이력을 생성합니다.
func NewProgramUpdateHistory(deviceName, deviceIP, programName, programVer string) *DeployHistory {
	return &DeployHistory{
		Type:        HistoryTypeProgram,
		Timestamp:   utils.Now(),
		DeviceName:  deviceName,
		DeviceIP:    deviceIP,
		ProgramName: programName,
		ProgramVer:  programVer,
		Status:      DeployStatusUnknown,
		Results:     []RuleResult{},
	}
}

// 이력 유형에 따른 표시 텍스트를 반환합니다.
func GetHistoryTypeText(historyType string) string {
	switch historyType {
	case HistoryTypeFirewall:
		return "방화벽 룰"
	case HistoryTypeProgram:
		return "패키지"
	default:
		return "방화벽 룰" // 기존 데이터 호환성
	}
}

// 규칙 결과를 추가합니다.
func (h *DeployHistory) AddResult(rule, status, reason string) {
	h.Results = append(h.Results, RuleResult{
		Rule:   rule,
		Status: status,
		Reason: reason,
	})
}

// 규칙 결과들을 기반으로 전체 상태를 계산합니다.
func (h *DeployHistory) CalculateStatus() {
	if len(h.Results) == 0 {
		h.Status = DeployStatusUnknown
		return
	}

	hasError := false
	for _, r := range h.Results {
		if r.Status != RuleStatusOK {
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

// 포맷된 시간 문자열을 반환합니다.
func (h *DeployHistory) GetTimestampString() string {
	return h.Timestamp.Time().Format("2006-01-02 15:04:05")
}

// 규칙 상태 코드를 표시 텍스트로 변환합니다.
func GetRuleStatusText(status string) string {
	// 대소문자 구분 없이 비교
	switch strings.ToLower(status) {
	case RuleStatusOK:
		return "OK"
	case RuleStatusError, RuleStatusUnfind, RuleStatusValidation:
		return "실패"
	case RuleStatusWrite:
		return "진행중"
	default:
		return status // 원본 상태 표시
	}
}

// 사유를 표시용 텍스트로 변환합니다.
func GetReasonText(reason string) string {
	// OK 또는 빈 문자열이면 "-" 표시
	if reason == "" || strings.ToLower(reason) == "ok" {
		return "-"
	}
	return reason
}
