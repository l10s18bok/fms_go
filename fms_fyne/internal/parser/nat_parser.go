package parser

import (
	"fmt"
	"strings"

	"fms/internal/model"
)

// ParseNATLine NAT 규칙 라인을 파싱하여 NATRule로 변환
// Smartfw 문서 형식: agent -m=insert -s=192.168.0.0/24 -p="ANY?SNAT" --dest=203.0.113.10 -a=NAT
func ParseNATLine(line string) (*model.NATRule, error) {
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

	// NAT 규칙 확인 (-a=NAT)
	if !strings.Contains(line, "-a=NAT") {
		return nil, fmt.Errorf("NAT 규칙이 아닙니다: %s", line)
	}

	rule := model.NewNATRule()
	parts := strings.Fields(line)

	for _, part := range parts {
		switch {
		case strings.HasPrefix(part, "-p="):
			// 문서 형식: -p="TCP?DNAT" 또는 -p="ANY?SNAT"
			protoStr := strings.Trim(part[3:], "\"")
			if idx := strings.Index(protoStr, "?"); idx != -1 {
				rule.Protocol = model.StringToProtocol(protoStr[:idx])
				rule.NATType = model.StringToNATType(protoStr[idx+1:])
			} else {
				rule.Protocol = model.StringToProtocol(protoStr)
			}
		case strings.HasPrefix(part, "--dport="):
			rule.MatchPort = part[8:]
		case strings.HasPrefix(part, "-s="):
			rule.MatchIP = part[3:]
		case strings.HasPrefix(part, "--dest="):
			// 문서 형식: --dest=IP 또는 --dest=IP:PORT
			dest := part[7:]
			if idx := strings.LastIndex(dest, ":"); idx != -1 {
				rule.TranslateIP = dest[:idx]
				rule.TranslatePort = dest[idx+1:]
			} else {
				rule.TranslateIP = dest
			}
		case strings.HasPrefix(part, "-i="):
			rule.InInterface = part[3:]
		case strings.HasPrefix(part, "-o="):
			rule.OutInterface = part[3:]
		}
	}

	return rule, nil
}

// NATRuleToLine NATRule을 agent 명령어 형식으로 변환
// 문서 형식: agent -m=insert -s=192.168.0.0/24 -p="ANY?SNAT" --dest=203.0.113.10 -a=NAT
func NATRuleToLine(rule *model.NATRule) string {
	if rule == nil {
		return ""
	}

	var parts []string
	parts = append(parts, "agent")
	parts = append(parts, "-m=insert")

	// 프로토콜?NAT타입 형식: -p="TCP?DNAT" 또는 -p="ANY?SNAT"
	protoStr := strings.ToUpper(model.ProtocolToString(rule.Protocol))
	natTypeStr := model.NATTypeToString(rule.NATType)
	parts = append(parts, fmt.Sprintf("-p=\"%s?%s\"", protoStr, natTypeStr))

	switch rule.NATType {
	case model.NATTypeDNAT:
		if rule.MatchIP != "" && rule.MatchIP != "ANY" {
			parts = append(parts, fmt.Sprintf("-s=%s", rule.MatchIP))
		}
		// --dest=IP:PORT
		if rule.TranslateIP != "" {
			dest := rule.TranslateIP
			if rule.TranslatePort != "" {
				dest += ":" + rule.TranslatePort
			}
			parts = append(parts, fmt.Sprintf("--dest=%s", dest))
		}
		if rule.MatchPort != "" {
			parts = append(parts, fmt.Sprintf("--dport=%s", rule.MatchPort))
		}

	case model.NATTypeSNAT:
		if rule.MatchIP != "" {
			parts = append(parts, fmt.Sprintf("-s=%s", rule.MatchIP))
		}
		if rule.TranslateIP != "" {
			parts = append(parts, fmt.Sprintf("--dest=%s", rule.TranslateIP))
		}
		if rule.OutInterface != "" {
			parts = append(parts, fmt.Sprintf("-o=%s", rule.OutInterface))
		}

	case model.NATTypeMASQUERADE:
		if rule.MatchIP != "" {
			parts = append(parts, fmt.Sprintf("-s=%s", rule.MatchIP))
		}
		if rule.OutInterface != "" {
			parts = append(parts, fmt.Sprintf("-o=%s", rule.OutInterface))
		}
	}

	// 공통: Action=NAT
	parts = append(parts, "-a=NAT")

	return strings.Join(parts, " ")
}

// NATRuleToSmartfw NATRule을 smartfw 형식으로 변환
// DNAT: req|INSERT|{ID}|ANY|NAT|{MatchIP}|{Proto}?DNAT|{TransIP}|{MatchPort},{TransPort}|{InIF}|{OutIF}
// SNAT: req|INSERT|{ID}|ANY|NAT|{MatchIP}|{Proto}?SNAT|{TransIP}|{Ports}|{InIF}|{OutIF}
func NATRuleToSmartfw(rule *model.NATRule, id string) string {
	if rule == nil {
		return ""
	}

	protoStr := strings.ToUpper(model.ProtocolToString(rule.Protocol))

	switch rule.NATType {
	case model.NATTypeDNAT:
		matchIP := rule.MatchIP
		if matchIP == "" {
			matchIP = "ANY"
		}
		ports := fmt.Sprintf("%s,%s", rule.MatchPort, rule.TranslatePort)
		return fmt.Sprintf("req|INSERT|%s|ANY|NAT|%s|%s?DNAT|%s|%s|%s|%s",
			id,
			matchIP,
			protoStr,
			rule.TranslateIP,
			ports,
			rule.InInterface,
			rule.OutInterface,
		)

	case model.NATTypeSNAT:
		matchIP := rule.MatchIP
		if matchIP == "" {
			matchIP = "ANY"
		}
		translateIP := rule.TranslateIP
		if translateIP == "" {
			translateIP = "ANY"
		}
		ports := rule.MatchPort
		if ports == "" {
			ports = "ANY"
		}
		return fmt.Sprintf("req|INSERT|%s|ANY|NAT|%s|%s?SNAT|%s|%s|%s|%s",
			id,
			matchIP,
			protoStr,
			translateIP,
			ports,
			rule.InInterface,
			rule.OutInterface,
		)

	case model.NATTypeMASQUERADE:
		matchIP := rule.MatchIP
		if matchIP == "" {
			matchIP = "ANY"
		}
		return fmt.Sprintf("req|INSERT|%s|ANY|NAT|%s|%s?MASQUERADE|ANY|ANY|%s|%s",
			id,
			matchIP,
			protoStr,
			rule.InInterface,
			rule.OutInterface,
		)
	}

	return ""
}

// ParseTextToNATRules 전체 텍스트에서 NAT 규칙 추출
func ParseTextToNATRules(text string) ([]*model.NATRule, []string, []error) {
	var rules []*model.NATRule
	var comments []string
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

		// NAT 규칙만 파싱 (IsNATLine 사용)
		if !IsNATLine(line) {
			continue
		}

		rule, err := ParseNATLine(line)
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

// NATRulesToText NAT 규칙 목록을 텍스트로 변환
func NATRulesToText(rules []*model.NATRule, comments []string) string {
	var lines []string

	// 주석 먼저 추가
	lines = append(lines, comments...)

	// 규칙 추가
	for _, rule := range rules {
		if line := NATRuleToLine(rule); line != "" {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// IsNATLine 라인이 NAT 규칙인지 확인
// Smartfw 문서 형식: -a=NAT 또는 -p="...?SNAT" / -p="...?DNAT" / -p="...?MASQUERADE"
func IsNATLine(line string) bool {
	return strings.Contains(line, "-a=NAT") ||
		strings.Contains(line, "?SNAT") ||
		strings.Contains(line, "?DNAT") ||
		strings.Contains(line, "?MASQUERADE")
}
