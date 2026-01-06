package parser

import (
	"fmt"
	"strings"

	"fms/internal/model"
)

// ParseLine 단일 라인을 파싱하여 FirewallRule로 변환
// 빈 줄이나 주석은 nil을 반환
func ParseLine(line string) (*model.FirewallRule, error) {
	line = strings.TrimSpace(line)

	// 빈 줄 처리
	if line == "" {
		return nil, nil
	}

	// 주석 라인 처리
	if strings.HasPrefix(line, "#") {
		return nil, nil
	}

	// agent 형식이 아니면 오류
	if !strings.HasPrefix(line, "agent ") {
		return nil, fmt.Errorf("알 수 없는 형식: %s", line)
	}

	rule := model.NewFirewallRule()
	parts := strings.Fields(line)

	for _, part := range parts {
		switch {
		case strings.HasPrefix(part, "-c="):
			rule.Chain = model.StringToChain(part[3:])
		case strings.HasPrefix(part, "-p="):
			rule.Protocol = model.StringToProtocol(part[3:])
		case strings.HasPrefix(part, "-a="):
			rule.Action = model.StringToAction(part[3:])
		case strings.HasPrefix(part, "--dport="):
			rule.DPort = part[8:]
		case strings.HasPrefix(part, "--sip="):
			rule.SIP = part[6:]
		case strings.HasPrefix(part, "--dip="):
			rule.DIP = part[6:]
		case part == "--black":
			rule.Black = true
		case part == "--white":
			rule.White = true
		}
	}

	return rule, nil
}

// RuleToLine FirewallRule을 텍스트 라인으로 변환
func RuleToLine(rule *model.FirewallRule) string {
	if rule == nil {
		return ""
	}

	var parts []string
	parts = append(parts, "agent")
	parts = append(parts, "-m=insert")
	parts = append(parts, fmt.Sprintf("-c=%s", model.ChainToString(rule.Chain)))
	parts = append(parts, fmt.Sprintf("-p=%s", model.ProtocolToString(rule.Protocol)))
	parts = append(parts, fmt.Sprintf("-a=%s", model.ActionToString(rule.Action)))

	// 선택 필드 (값이 있을 때만 출력)
	if rule.DPort != "" {
		parts = append(parts, fmt.Sprintf("--dport=%s", rule.DPort))
	}
	if rule.SIP != "" {
		parts = append(parts, fmt.Sprintf("--sip=%s", rule.SIP))
	}
	if rule.DIP != "" {
		parts = append(parts, fmt.Sprintf("--dip=%s", rule.DIP))
	}

	// 플래그 (true일 때만 출력)
	if rule.Black {
		parts = append(parts, "--black")
	}
	if rule.White {
		parts = append(parts, "--white")
	}

	return strings.Join(parts, " ")
}

// ParseTextToRules 전체 텍스트를 파싱하여 규칙 목록으로 변환
// 주석과 빈 줄은 별도로 보존
func ParseTextToRules(text string) ([]*model.FirewallRule, []string, []error) {
	var rules []*model.FirewallRule
	var comments []string // 주석 라인 보존
	var errors []error

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// 빈 줄 무시
		if line == "" {
			continue
		}

		// 주석 라인 보존
		if strings.HasPrefix(line, "#") {
			comments = append(comments, line)
			continue
		}

		rule, err := ParseLine(line)
		if err != nil {
			errors = append(errors, fmt.Errorf("라인 %d: %w", i+1, err))
			continue
		}

		if rule != nil {
			rules = append(rules, rule)
		}
	}

	return rules, comments, errors
}

// RulesToText 규칙 목록을 텍스트로 변환
func RulesToText(rules []*model.FirewallRule, comments []string) string {
	var lines []string

	// 주석 먼저 추가
	lines = append(lines, comments...)

	// 규칙 추가
	for _, rule := range rules {
		if line := RuleToLine(rule); line != "" {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}
