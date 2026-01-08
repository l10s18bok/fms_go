package parser

import (
	"fmt"
	"strings"

	"fms_wails/internal/model"
)

// ParseNATLine NAT 규칙 라인을 파싱하여 NATRule로 변환
// agent -m=insert -t=nat --nat-type=dnat -p=tcp --match-port=6080 --to-dest=192.168.30.180:8080
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

	// -t=nat 확인
	if !strings.Contains(line, "-t=nat") {
		return nil, fmt.Errorf("NAT 규칙이 아닙니다: %s", line)
	}

	rule := model.NewNATRule()
	parts := strings.Fields(line)

	for _, part := range parts {
		switch {
		case strings.HasPrefix(part, "--nat-type="):
			rule.NATType = model.StringToNATType(part[11:])
		case strings.HasPrefix(part, "-p="):
			rule.Protocol = model.StringToProtocol(part[3:])
		case strings.HasPrefix(part, "--match-port="):
			rule.MatchPort = part[13:]
		case strings.HasPrefix(part, "--match-ip="):
			rule.MatchIP = part[11:]
		case strings.HasPrefix(part, "-s="):
			rule.MatchIP = part[3:]
		case strings.HasPrefix(part, "--to-dest="):
			// 192.168.30.180:8080 형식 파싱
			dest := part[10:]
			if idx := strings.LastIndex(dest, ":"); idx != -1 {
				rule.TranslateIP = dest[:idx]
				rule.TranslatePort = dest[idx+1:]
			} else {
				rule.TranslateIP = dest
			}
		case strings.HasPrefix(part, "--to-source="):
			rule.TranslateIP = part[12:]
		case strings.HasPrefix(part, "-i="):
			rule.InInterface = part[3:]
		case strings.HasPrefix(part, "-o="):
			rule.OutInterface = part[3:]
		case strings.HasPrefix(part, "--desc="):
			rule.Description = part[7:]
		}
	}

	return rule, nil
}

// NATRuleToLine NATRule을 agent 명령어 형식으로 변환
func NATRuleToLine(rule *model.NATRule) string {
	if rule == nil {
		return ""
	}

	var parts []string
	parts = append(parts, "agent")
	parts = append(parts, "-m=insert")
	parts = append(parts, "-t=nat")
	parts = append(parts, fmt.Sprintf("--nat-type=%s", strings.ToLower(model.NATTypeToString(rule.NATType))))
	parts = append(parts, fmt.Sprintf("-p=%s", model.ProtocolToString(rule.Protocol)))

	switch rule.NATType {
	case model.NATTypeDNAT:
		if rule.MatchPort != "" {
			parts = append(parts, fmt.Sprintf("--match-port=%s", rule.MatchPort))
		}
		if rule.MatchIP != "" && rule.MatchIP != "ANY" {
			parts = append(parts, fmt.Sprintf("-s=%s", rule.MatchIP))
		}
		// --to-dest=IP:PORT
		if rule.TranslateIP != "" {
			dest := rule.TranslateIP
			if rule.TranslatePort != "" {
				dest += ":" + rule.TranslatePort
			}
			parts = append(parts, fmt.Sprintf("--to-dest=%s", dest))
		}

	case model.NATTypeSNAT:
		if rule.MatchIP != "" {
			parts = append(parts, fmt.Sprintf("-s=%s", rule.MatchIP))
		}
		if rule.TranslateIP != "" {
			parts = append(parts, fmt.Sprintf("--to-source=%s", rule.TranslateIP))
		}
		if rule.InInterface != "" {
			parts = append(parts, fmt.Sprintf("-i=%s", rule.InInterface))
		}
		if rule.OutInterface != "" {
			parts = append(parts, fmt.Sprintf("-o=%s", rule.OutInterface))
		}

	case model.NATTypeMASQUERADE:
		if rule.MatchIP != "" {
			parts = append(parts, fmt.Sprintf("-s=%s", rule.MatchIP))
		}
		if rule.InInterface != "" {
			parts = append(parts, fmt.Sprintf("-i=%s", rule.InInterface))
		}
		if rule.OutInterface != "" {
			parts = append(parts, fmt.Sprintf("-o=%s", rule.OutInterface))
		}
	}

	if rule.Description != "" {
		parts = append(parts, fmt.Sprintf("--desc=%s", rule.Description))
	}

	return strings.Join(parts, " ")
}

// NATRuleToSmartfw NATRule을 smartfw 형식으로 변환
// DNAT: req|INSERT|{ID}|ANY|NAT|{SRC}|{PROTOCOL}?DNAT|{DEST_IP}|{MATCH_PORT},{TRANSLATE_PORT}|{IN_IF}|{OUT_IF}
// SNAT: req|INSERT|{ID}|ANY|NAT|{SRC}|{PROTOCOL}?SNAT|{DEST}|{PORTS}|{IN_IF}|{OUT_IF}
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

		// NAT 규칙만 파싱 (-t=nat 포함된 라인)
		if !strings.Contains(line, "-t=nat") {
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
func IsNATLine(line string) bool {
	return strings.Contains(line, "-t=nat")
}
