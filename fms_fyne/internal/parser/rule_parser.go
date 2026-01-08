package parser

import (
	"fmt"
	"strings"

	"fms/internal/model"
)

// ParseProtocolWithOptions 프로토콜 문자열을 파싱
// 입력: "tcp?flags=syn/syn" 또는 "tcp"
// 출력: Protocol, *ProtocolOptions, error
func ParseProtocolWithOptions(s string) (model.Protocol, *model.ProtocolOptions, error) {
	// "?" 기준으로 분리
	parts := strings.SplitN(s, "?", 2)
	protocol := model.StringToProtocol(parts[0])

	if len(parts) == 1 {
		// 옵션 없음
		return protocol, nil, nil
	}

	// 쿼리 스트링 파싱
	opts := &model.ProtocolOptions{}
	params := strings.Split(parts[1], "&")

	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "flags":
			opts.TCPFlags = kv[1]
		case "type":
			opts.ICMPType = kv[1]
		case "code":
			opts.ICMPCode = kv[1]
		}
	}

	return protocol, opts, nil
}

// FormatProtocolWithOptions 프로토콜과 옵션을 문자열로 변환
// 입력: Protocol=TCP, Options={TCPFlags: "syn/syn"}
// 출력: "tcp?flags=syn/syn"
func FormatProtocolWithOptions(p model.Protocol, opts *model.ProtocolOptions) string {
	base := model.ProtocolToString(p)

	if opts == nil || opts.IsEmpty() {
		return base
	}

	var params []string

	// TCP flags
	if opts.TCPFlags != "" {
		params = append(params, "flags="+opts.TCPFlags)
	}

	// ICMP type
	if opts.ICMPType != "" {
		params = append(params, "type="+opts.ICMPType)
	}

	// ICMP code
	if opts.ICMPCode != "" {
		params = append(params, "code="+opts.ICMPCode)
	}

	if len(params) == 0 {
		return base
	}

	return base + "?" + strings.Join(params, "&")
}

// FormatOptionsOnly 옵션만 문자열로 변환 (프로토콜 제외)
// 입력: Options={TCPFlags: "syn/syn"}
// 출력: "flags=syn/syn"
func FormatOptionsOnly(opts *model.ProtocolOptions) string {
	if opts == nil || opts.IsEmpty() {
		return ""
	}

	var params []string

	// TCP flags
	if opts.TCPFlags != "" {
		params = append(params, "flags="+opts.TCPFlags)
	}

	// ICMP type
	if opts.ICMPType != "" {
		params = append(params, "type="+opts.ICMPType)
	}

	// ICMP code
	if opts.ICMPCode != "" {
		params = append(params, "code="+opts.ICMPCode)
	}

	return strings.Join(params, "&")
}

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
			// 프로토콜 옵션 파싱 (쿼리 스트링 형식 지원)
			proto, opts, _ := ParseProtocolWithOptions(part[3:])
			rule.Protocol = proto
			rule.Options = opts
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
	// 프로토콜 옵션 포함하여 포맷
	parts = append(parts, fmt.Sprintf("-p=%s", FormatProtocolWithOptions(rule.Protocol, rule.Options)))
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

		// NAT 규칙은 건너뛰기 (ParseTextToNATRules에서 처리)
		if strings.Contains(line, "-t=nat") {
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
