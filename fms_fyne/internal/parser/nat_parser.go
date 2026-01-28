package parser

import (
	"fmt"
	"strings"

	"fms/internal/model"
)

// ParseNATLine NAT 규칙 라인을 파싱하여 NATRule로 변환
// 지원 형식: ./agent -f -I -a NAT -p TCP?SNAT -s 192.168.1.0/24 -e 1.2.3.4 -o eth0
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

	// agent 형식 확인 (./agent)
	if !strings.HasPrefix(line, "./agent ") {
		return nil, fmt.Errorf("알 수 없는 형식: %s", line)
	}

	// NAT 규칙 확인
	if !IsNATLine(line) {
		return nil, fmt.Errorf("NAT 규칙이 아닙니다: %s", line)
	}

	rule := model.NewNATRule()
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
				protoStr := parts[i]
				// TCP?SNAT, TCP?DNAT, TCP?MASQUERADE 형식 파싱
				rule.Protocol, rule.NATType = parseProtocolWithNATType(protoStr)
			}
		case "-a":
			if i+1 < len(parts) {
				i++
				// NAT 액션은 무시 (이미 NAT 규칙임을 확인함)
			}
		case "-n":
			if i+1 < len(parts) {
				i++
				rule.MatchPort = parts[i]
			}
		case "-s":
			if i+1 < len(parts) {
				i++
				rule.MatchIP = parts[i]
			}
		case "-e":
			if i+1 < len(parts) {
				i++
				dest := parts[i]
				if idx := strings.LastIndex(dest, ":"); idx != -1 {
					rule.TranslateIP = dest[:idx]
					rule.TranslatePort = dest[idx+1:]
				} else {
					rule.TranslateIP = dest
				}
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

// parseProtocolWithNATType 프로토콜 문자열에서 NAT 타입 추출
// 입력: "TCP?SNAT", "UDP?DNAT", "TCP?MASQUERADE", "TCP"
// 출력: Protocol, NATType
func parseProtocolWithNATType(s string) (model.Protocol, model.NATType) {
	s = strings.ToUpper(s)

	// "?" 기준으로 분리
	if idx := strings.Index(s, "?"); idx != -1 {
		protoStr := s[:idx]
		natTypeStr := s[idx+1:]
		return model.StringToProtocol(protoStr), model.StringToNATType(natTypeStr)
	}

	// NAT 타입이 없으면 기본값 DNAT
	return model.StringToProtocol(s), model.NATTypeDNAT
}

// NATRuleToLine NATRule을 agent 명령어 형식으로 변환
// 출력 형식: ./agent -f -I -a NAT -p TCP?SNAT -s 192.168.1.0/24 -e 1.2.3.4 -o eth0
func NATRuleToLine(rule *model.NATRule) string {
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

	// Action: -a NAT
	parts = append(parts, "-a")
	parts = append(parts, "NAT")

	// Protocol with NAT Type: -p TCP?SNAT
	parts = append(parts, "-p")
	protoStr := strings.ToUpper(model.ProtocolToString(rule.Protocol))
	natTypeStr := model.NATTypeToString(rule.NATType)
	parts = append(parts, protoStr+"?"+natTypeStr)

	// Port: -n port (DNAT에서 주로 사용)
	if rule.MatchPort != "" {
		parts = append(parts, "-n")
		parts = append(parts, rule.MatchPort)
	}

	// Source IP: -s ip (MatchIP)
	if rule.MatchIP != "" && rule.MatchIP != "ANY" {
		parts = append(parts, "-s")
		parts = append(parts, rule.MatchIP)
	}

	// Destination IP: -e ip (TranslateIP:TranslatePort)
	if rule.TranslateIP != "" {
		parts = append(parts, "-e")
		if rule.TranslatePort != "" {
			parts = append(parts, rule.TranslateIP+":"+rule.TranslatePort)
		} else {
			parts = append(parts, rule.TranslateIP)
		}
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
// 지원 형식: -a NAT 또는 -a nat
func IsNATLine(line string) bool {
	upperLine := strings.ToUpper(line)
	return strings.Contains(upperLine, "-A NAT")
}
