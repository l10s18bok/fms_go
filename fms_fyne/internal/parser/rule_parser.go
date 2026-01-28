package parser

import (
	"fmt"
	"strings"

	"fms/internal/model"
)

// ParseProtocolWithOptions 프로토콜 문자열을 파싱
// 입력: "tcp?flags=syn/syn", "IPS?syn-flood&limit=50&seconds=1&enable=1"
// 출력: Protocol, *ProtocolOptions, error
func ParseProtocolWithOptions(s string) (model.Protocol, *model.ProtocolOptions, error) {
	// 따옴표 제거
	s = strings.Trim(s, "\"")

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

	// IPS 프로토콜인 경우: IPS?syn-flood&limit=50&seconds=1&enable=1
	// 첫 번째 파라미터가 key=value 형식이 아니면 IPSType으로 처리
	if protocol == model.ProtocolIPS && len(params) > 0 {
		firstParam := params[0]
		if !strings.Contains(firstParam, "=") {
			// 첫 번째 파라미터가 IPS 타입 (예: syn-flood)
			opts.IPSType = parts[1] // 전체 옵션 문자열 저장 (syn-flood&limit=50&...)
			return protocol, opts, nil
		}
	}

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
// 입력: Protocol=IPS, Options={IPSType: "syn-flood&limit=50&seconds=1&enable=1"}
// 출력: "\"IPS?syn-flood&limit=50&seconds=1&enable=1\""
func FormatProtocolWithOptions(p model.Protocol, opts *model.ProtocolOptions) string {
	base := model.ProtocolToString(p)

	if opts == nil || opts.IsEmpty() {
		return base
	}

	// IPS 프로토콜인 경우 특수 처리
	if p == model.ProtocolIPS && opts.IPSType != "" {
		return fmt.Sprintf("\"%s?%s\"", base, opts.IPSType)
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

	// IPS 옵션
	if opts.IPSType != "" {
		return opts.IPSType
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

// ParseOptionsOnly 옵션 문자열만 파싱 (프로토콜 제외)
// 입력: "flags=syn/syn" 또는 "type=8&code=0"
// 출력: *ProtocolOptions
func ParseOptionsOnly(s string) *model.ProtocolOptions {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return nil
	}

	opts := &model.ProtocolOptions{}
	params := strings.Split(s, "&")

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

	if opts.IsEmpty() {
		return nil
	}
	return opts
}

// ParseLine 단일 라인을 파싱하여 FirewallRule로 변환
// 빈 줄이나 주석은 nil을 반환
// 지원 형식: ./agent -f -I -c INPUT -a ACCEPT -p TCP -n 80
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

	// agent 형식 확인 (./agent)
	if !strings.HasPrefix(line, "./agent ") {
		return nil, fmt.Errorf("알 수 없는 형식: %s", line)
	}

	rule := model.NewFirewallRule()
	parts := strings.Fields(line)

	// 인덱스 기반 파싱 (공백 구분 형식)
	for i := 0; i < len(parts); i++ {
		part := parts[i]

		switch part {
		case "-f":
			// 방화벽 모드 플래그 - 무시
			continue
		case "-I":
			rule.Command = model.CommandInsert
		case "-D":
			rule.Command = model.CommandDelete
		case "-c":
			if i+1 < len(parts) {
				i++
				rule.Chain = model.StringToChain(parts[i])
			}
		case "-p":
			if i+1 < len(parts) {
				i++
				proto, opts, _ := ParseProtocolWithOptions(parts[i])
				rule.Protocol = proto
				rule.Options = opts
			}
		case "-a":
			if i+1 < len(parts) {
				i++
				actionStr := parts[i]
				switch actionStr {
				case "blocklist":
					rule.Black = true
				case "whitelist":
					rule.White = true
				default:
					rule.Action = model.StringToAction(actionStr)
				}
			}
		case "-n":
			if i+1 < len(parts) {
				i++
				rule.DPort = parts[i]
			}
		case "-s":
			if i+1 < len(parts) {
				i++
				rule.SIP = parts[i]
			}
		case "-e":
			if i+1 < len(parts) {
				i++
				rule.DIP = parts[i]
			}
		case "-i":
			if i+1 < len(parts) {
				i++
				rule.InInterface = parts[i]
			}
		case "-o":
			if i+1 < len(parts) {
				i++
				rule.OutInterface = parts[i]
			}
		}
	}

	return rule, nil
}

// RuleToLine FirewallRule을 텍스트 라인으로 변환
// 출력 형식: ./agent -f -I -c INPUT -a ACCEPT -p TCP -n 80
func RuleToLine(rule *model.FirewallRule) string {
	if rule == nil {
		return ""
	}

	var parts []string
	parts = append(parts, "./agent")
	parts = append(parts, "-f")

	// Command (Insert/Delete)
	if rule.Command == model.CommandDelete {
		parts = append(parts, "-D")
	} else {
		parts = append(parts, "-I")
	}

	// ACL 규칙 (blocklist/whitelist)은 별도 형식
	if rule.Black || rule.White {
		// 소스 IP
		if rule.SIP != "" {
			parts = append(parts, "-s")
			parts = append(parts, rule.SIP)
		}
		// ACL Action
		if rule.Black {
			parts = append(parts, "-a")
			parts = append(parts, "blocklist")
		} else if rule.White {
			parts = append(parts, "-a")
			parts = append(parts, "whitelist")
		}
		return strings.Join(parts, " ")
	}

	// IPS 규칙은 체인이 없음
	if rule.Protocol == model.ProtocolIPS {
		parts = append(parts, "-p")
		parts = append(parts, FormatProtocolWithOptions(rule.Protocol, rule.Options))
		parts = append(parts, "-a")
		parts = append(parts, model.ActionToString(rule.Action))
		return strings.Join(parts, " ")
	}

	// 일반 방화벽 규칙: -c Chain
	parts = append(parts, "-c")
	parts = append(parts, model.ChainToString(rule.Chain))

	// Action: -a Action
	parts = append(parts, "-a")
	parts = append(parts, model.ActionToString(rule.Action))

	// Protocol: -p Protocol
	parts = append(parts, "-p")
	parts = append(parts, strings.ToUpper(FormatProtocolWithOptions(rule.Protocol, rule.Options)))

	// 선택 필드 (값이 있을 때만 출력)
	// Port: -n port
	if rule.DPort != "" {
		parts = append(parts, "-n")
		parts = append(parts, rule.DPort)
	}
	// Source IP: -s ip
	if rule.SIP != "" {
		parts = append(parts, "-s")
		parts = append(parts, rule.SIP)
	}
	// Destination IP: -e ip
	if rule.DIP != "" {
		parts = append(parts, "-e")
		parts = append(parts, rule.DIP)
	}
	// In Interface: -i interface
	if rule.InInterface != "" {
		parts = append(parts, "-i")
		parts = append(parts, rule.InInterface)
	}
	// Out Interface: -o interface
	if rule.OutInterface != "" {
		parts = append(parts, "-o")
		parts = append(parts, rule.OutInterface)
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
		if IsNATLine(line) {
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

// IsIPSLine 라인이 IPS 규칙인지 확인
func IsIPSLine(line string) bool {
	// IPS 형식: -p="IPS?syn-flood&limit=50&seconds=1&enable=1"
	return strings.Contains(line, "-p=\"IPS?") || strings.Contains(line, "-p=IPS?")
}
