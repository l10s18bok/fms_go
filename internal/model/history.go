package model

import "fms/internal/utils"

// 배포 이력을 나타냅니다.
type DeployHistory struct {
	ID          int            `json:"id"`              // 고유 ID (Auto Increment)
	Timestamp   utils.JSONTime `json:"timestamp"`       // 배포 시간
	DeviceIP    string       `json:"deviceIp"`        // 장비 IP
	TemplateVer string       `json:"templateVersion"` // 배포한 템플릿 버전
	Status      string       `json:"status"`          // 배포 상태 (success/fail/error)
	Results     []RuleResult `json:"results"`         // 규칙별 결과
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

// 새로운 배포 이력을 생성합니다.
func NewDeployHistory(deviceIP, templateVer string) *DeployHistory {
	return &DeployHistory{
		Timestamp:   utils.Now(),
		DeviceIP:    deviceIP,
		TemplateVer: templateVer,
		Status:      DeployStatusUnknown,
		Results:     []RuleResult{},
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
	switch status {
	case RuleStatusOK:
		return "성공"
	case RuleStatusError, RuleStatusUnfind, RuleStatusValidation:
		return "실패"
	case RuleStatusWrite:
		return "진행중"
	default:
		return "-"
	}
}
