package model

import "strings"

// Chain 방화벽 체인 타입
type Chain int

const (
	ChainINPUT       Chain = 0
	ChainOUTPUT      Chain = 1
	ChainFORWARD     Chain = 2
	ChainPREROUTING  Chain = 3
	ChainPOSTROUTING Chain = 4
)

// Protocol 프로토콜 타입
type Protocol int

const (
	ProtocolTCP  Protocol = 6
	ProtocolUDP  Protocol = 17
	ProtocolICMP Protocol = 1
	ProtocolANY  Protocol = 255
)

// Action 규칙 액션 타입
type Action int

const (
	ActionDROP   Action = 0
	ActionACCEPT Action = 1
	ActionREJECT Action = 2
)

// FirewallRule 방화벽 규칙 구조체
type FirewallRule struct {
	Chain    Chain
	Protocol Protocol
	Action   Action
	DPort    string // Destination 포트
	SIP      string // Source IP (콤마리스트 지원)
	DIP      string // Destination IP (콤마리스트 지원)
	Black    bool   // 블랙리스트 규칙 여부
	White    bool   // 화이트리스트 규칙 여부
}

// NewFirewallRule 기본값으로 새 규칙 생성
func NewFirewallRule() *FirewallRule {
	return &FirewallRule{
		Chain:    ChainINPUT,
		Protocol: ProtocolTCP,
		Action:   ActionDROP,
	}
}

// ChainToString Chain을 문자열로 변환
func ChainToString(c Chain) string {
	switch c {
	case ChainINPUT:
		return "INPUT"
	case ChainOUTPUT:
		return "OUTPUT"
	case ChainFORWARD:
		return "FORWARD"
	case ChainPREROUTING:
		return "PREROUTING"
	case ChainPOSTROUTING:
		return "POSTROUTING"
	default:
		return "INPUT"
	}
}

// StringToChain 문자열을 Chain으로 변환
func StringToChain(s string) Chain {
	switch strings.ToUpper(s) {
	case "INPUT":
		return ChainINPUT
	case "OUTPUT":
		return ChainOUTPUT
	case "FORWARD":
		return ChainFORWARD
	case "PREROUTING":
		return ChainPREROUTING
	case "POSTROUTING":
		return ChainPOSTROUTING
	default:
		return ChainINPUT
	}
}

// ProtocolToString Protocol을 문자열로 변환
func ProtocolToString(p Protocol) string {
	switch p {
	case ProtocolTCP:
		return "tcp"
	case ProtocolUDP:
		return "udp"
	case ProtocolICMP:
		return "icmp"
	case ProtocolANY:
		return "any"
	default:
		return "tcp"
	}
}

// StringToProtocol 문자열을 Protocol로 변환
func StringToProtocol(s string) Protocol {
	switch strings.ToLower(s) {
	case "tcp":
		return ProtocolTCP
	case "udp":
		return ProtocolUDP
	case "icmp":
		return ProtocolICMP
	case "any":
		return ProtocolANY
	default:
		return ProtocolTCP
	}
}

// ActionToString Action을 문자열로 변환
func ActionToString(a Action) string {
	switch a {
	case ActionDROP:
		return "DROP"
	case ActionACCEPT:
		return "ACCEPT"
	case ActionREJECT:
		return "REJECT"
	default:
		return "DROP"
	}
}

// StringToAction 문자열을 Action으로 변환
func StringToAction(s string) Action {
	switch strings.ToUpper(s) {
	case "DROP":
		return ActionDROP
	case "ACCEPT":
		return ActionACCEPT
	case "REJECT":
		return ActionREJECT
	default:
		return ActionDROP
	}
}

// GetChainOptions UI Select용 Chain 옵션 목록
func GetChainOptions() []string {
	return []string{"INPUT", "OUTPUT", "FORWARD"}
	// "PREROUTING", "POSTROUTING" - 현재 미사용
}

// GetProtocolOptions UI Select용 Protocol 옵션 목록
func GetProtocolOptions() []string {
	return []string{"tcp", "udp", "icmp", "any"}
}

// GetActionOptions UI Select용 Action 옵션 목록
func GetActionOptions() []string {
	return []string{"DROP", "ACCEPT", "REJECT"}
}
