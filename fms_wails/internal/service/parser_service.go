package service

import (
	"encoding/json"

	"fms_wails/internal/model"
	"fms_wails/internal/parser"
)

// ParseRulesResult는 규칙 파싱 결과입니다.
type ParseRulesResult struct {
	Rules    []*model.FirewallRule `json:"rules"`
	Comments []string              `json:"comments"`
	Errors   []string              `json:"errors"`
}

// ParseNATRulesResult는 NAT 규칙 파싱 결과입니다.
type ParseNATRulesResult struct {
	Rules    []*model.NATRule `json:"rules"`
	Comments []string         `json:"comments"`
	Errors   []string         `json:"errors"`
}

// ParserService는 규칙 파싱 관련 비즈니스 로직을 처리합니다.
type ParserService struct{}

// NewParserService는 새로운 ParserService를 생성합니다.
func NewParserService() *ParserService {
	return &ParserService{}
}

// ParseRules는 텍스트를 규칙 배열로 파싱합니다.
func (s *ParserService) ParseRules(text string) *ParseRulesResult {
	rules, comments, errors := parser.ParseTextToRules(text)

	var errorMessages []string
	for _, err := range errors {
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	return &ParseRulesResult{
		Rules:    rules,
		Comments: comments,
		Errors:   errorMessages,
	}
}

// RulesToText는 규칙 배열을 텍스트로 변환합니다.
func (s *ParserService) RulesToText(rulesJSON string, commentsJSON string) string {
	var rules []*model.FirewallRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return ""
	}

	var comments []string
	if commentsJSON != "" {
		if err := json.Unmarshal([]byte(commentsJSON), &comments); err != nil {
			comments = nil
		}
	}

	return parser.RulesToText(rules, comments)
}

// ParseNATRules는 텍스트를 NAT 규칙 배열로 파싱합니다.
func (s *ParserService) ParseNATRules(text string) *ParseNATRulesResult {
	rules, comments, errors := parser.ParseTextToNATRules(text)

	var errorMessages []string
	for _, err := range errors {
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	return &ParseNATRulesResult{
		Rules:    rules,
		Comments: comments,
		Errors:   errorMessages,
	}
}

// NATRulesToText는 NAT 규칙 배열을 텍스트로 변환합니다.
func (s *ParserService) NATRulesToText(rulesJSON string, commentsJSON string) string {
	var rules []*model.NATRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return ""
	}

	var comments []string
	if commentsJSON != "" {
		if err := json.Unmarshal([]byte(commentsJSON), &comments); err != nil {
			comments = nil
		}
	}

	return parser.NATRulesToText(rules, comments)
}

// GetChainOptions는 Chain 옵션 목록을 반환합니다.
func (s *ParserService) GetChainOptions() []string {
	return model.GetChainOptions()
}

// GetProtocolOptions는 Protocol 옵션 목록을 반환합니다.
func (s *ParserService) GetProtocolOptions() []string {
	return model.GetProtocolOptions()
}

// GetActionOptions는 Action 옵션 목록을 반환합니다.
func (s *ParserService) GetActionOptions() []string {
	return model.GetActionOptions()
}

// GetTCPFlagsPresets는 TCP Flags 프리셋 목록을 반환합니다.
func (s *ParserService) GetTCPFlagsPresets() []model.TCPFlagsPreset {
	return model.GetTCPFlagsPresets()
}

// GetTCPFlagsList는 개별 TCP Flags 목록을 반환합니다.
func (s *ParserService) GetTCPFlagsList() []string {
	return model.GetTCPFlagsList()
}

// GetICMPTypeOptions는 ICMP Type 옵션 목록을 반환합니다.
func (s *ParserService) GetICMPTypeOptions() []string {
	return model.GetICMPTypeOptions()
}

// GetICMPCodeOptions는 ICMP Code 옵션 목록을 반환합니다.
func (s *ParserService) GetICMPCodeOptions() []string {
	return model.GetICMPCodeOptions()
}

// NewFirewallRule은 기본값이 설정된 새 규칙을 생성합니다.
func (s *ParserService) NewFirewallRule() *model.FirewallRule {
	return model.NewFirewallRule()
}

// GetNATTypeOptions는 NAT 타입 옵션 목록을 반환합니다.
func (s *ParserService) GetNATTypeOptions() []string {
	return model.GetNATTypeOptions()
}

// GetSNATTypeOptions는 SNAT 폼용 타입 옵션을 반환합니다.
func (s *ParserService) GetSNATTypeOptions() []string {
	return model.GetSNATTypeOptions()
}

// NewNATRule은 기본값이 설정된 새 NAT 규칙을 생성합니다.
func (s *ParserService) NewNATRule() *model.NATRule {
	return model.NewNATRule()
}

// NewDNATRule은 DNAT 규칙을 생성합니다.
func (s *ParserService) NewDNATRule() *model.NATRule {
	return model.NewDNATRule()
}

// NewSNATRule은 SNAT 규칙을 생성합니다.
func (s *ParserService) NewSNATRule() *model.NATRule {
	return model.NewSNATRule()
}
