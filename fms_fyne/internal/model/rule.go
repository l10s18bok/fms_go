package model

import (
	"fmt"
	"strconv"
	"strings"
)

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

// ProtocolOptions 프로토콜별 세부 옵션
type ProtocolOptions struct {
	// TCP 옵션
	TCPFlags string // 예: "syn/syn", "syn,ack/syn"

	// ICMP 옵션
	ICMPType string // 예: "echo-request", "8"
	ICMPCode string // 예: "0", "3" (선택)
}

// IsEmpty 옵션이 비어있는지 확인
func (o *ProtocolOptions) IsEmpty() bool {
	if o == nil {
		return true
	}
	return o.TCPFlags == "" && o.ICMPType == "" && o.ICMPCode == ""
}

// HasTCPOptions TCP 옵션이 있는지 확인
func (o *ProtocolOptions) HasTCPOptions() bool {
	if o == nil {
		return false
	}
	return o.TCPFlags != ""
}

// HasICMPOptions ICMP 옵션이 있는지 확인
func (o *ProtocolOptions) HasICMPOptions() bool {
	if o == nil {
		return false
	}
	return o.ICMPType != "" || o.ICMPCode != ""
}

// FirewallRule 방화벽 규칙 구조체
type FirewallRule struct {
	Chain    Chain
	Protocol Protocol
	Options  *ProtocolOptions // 프로토콜 옵션
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
	return []string{"DROP", "ACCEPT"}
}

// TCPFlagsPreset TCP Flags 프리셋 정의
type TCPFlagsPreset struct {
	Name        string   // 프리셋 이름 (UI 표시용)
	MaskFlags   []string // 검사할 플래그
	SetFlags    []string // 설정된 플래그
	Description string   // 설명
}

// ToFlagsString 프리셋을 flags 문자열로 변환
// 예: "syn,rst,ack,fin/syn"
func (p *TCPFlagsPreset) ToFlagsString() string {
	if len(p.MaskFlags) == 0 {
		return ""
	}
	mask := strings.Join(p.MaskFlags, ",")
	set := strings.Join(p.SetFlags, ",")
	return mask + "/" + set
}

// GetTCPFlagsPresets 프리셋 목록 반환
func GetTCPFlagsPresets() []TCPFlagsPreset {
	return []TCPFlagsPreset{
		{
			Name:        "None",
			MaskFlags:   nil,
			SetFlags:    nil,
			Description: "Match all TCP packets",
		},
		{
			Name:        "New Connection (SYN)",
			MaskFlags:   []string{"syn", "rst", "ack", "fin"},
			SetFlags:    []string{"syn"},
			Description: "Match new connection requests",
		},
		{
			Name:        "Established (ACK)",
			MaskFlags:   []string{"ack"},
			SetFlags:    []string{"ack"},
			Description: "Match established connections",
		},
		{
			Name:        "Block NULL Scan",
			MaskFlags:   []string{"syn", "rst", "ack", "fin", "psh", "urg"},
			SetFlags:    nil,
			Description: "Block packets with no flags",
		},
		{
			Name:        "Block XMAS Scan",
			MaskFlags:   []string{"syn", "rst", "ack", "fin", "psh", "urg"},
			SetFlags:    []string{"fin", "psh", "urg"},
			Description: "Block abnormal flag combination",
		},
		{
			Name:        "Block SYN+FIN",
			MaskFlags:   []string{"syn", "fin"},
			SetFlags:    []string{"syn", "fin"},
			Description: "Block abnormal flag combination",
		},
		{
			Name:        "Block SYN+RST",
			MaskFlags:   []string{"syn", "rst"},
			SetFlags:    []string{"syn", "rst"},
			Description: "Block abnormal flag combination",
		},
		{
			Name:        "Block FIN+RST",
			MaskFlags:   []string{"fin", "rst"},
			SetFlags:    []string{"fin", "rst"},
			Description: "Block abnormal flag combination",
		},
		{
			Name:        "Custom",
			MaskFlags:   nil,
			SetFlags:    nil,
			Description: "Manual checkbox selection",
		},
	}
}

// FindPresetByFlags flags 문자열에 매칭되는 프리셋 찾기
// 매칭되는 프리셋 없으면 "Custom" 반환
func FindPresetByFlags(flags string) *TCPFlagsPreset {
	presets := GetTCPFlagsPresets()

	// 빈 문자열은 "None"
	if flags == "" {
		return &presets[0]
	}

	// 각 프리셋과 비교
	for i, preset := range presets {
		if preset.Name == "Custom" {
			continue
		}
		if preset.ToFlagsString() == flags {
			return &presets[i]
		}
	}

	// 매칭되지 않으면 Custom
	return &presets[len(presets)-1]
}

// GetTCPFlagsList TCP flags 옵션 목록 (체크박스용)
func GetTCPFlagsList() []string {
	return []string{"syn", "ack", "fin", "rst", "psh", "urg"}
}

// GetICMPTypeOptions ICMP type 옵션 목록 (UI Select용, smartfw 순서와 동일)
func GetICMPTypeOptions() []string {
	return []string{
		"None",
		"echo-reply (0)",
		"destination-unreachable (3)",
		"source-quench (4)",
		"echo-redirect (5)",
		"echo-request (8)",
		"time-exceeded (11)",
		"parameter-problem (12)",
		"timestamp-request (13)",
		"timestamp-reply (14)",
		"information-request (15)",
		"information-reply (16)",
		"addressmask-request (17)",
		"addressmask-reply (18)",
	}
}

// icmpTypeMap ICMP 타입 이름 → 숫자 매핑 (smartfw 동일)
var icmpTypeMap = map[string]int{
	"echo-reply":              0,
	"destination-unreachable": 3,
	"source-quench":           4,
	"echo-redirect":           5,
	"echo-request":            8,
	"time-exceeded":           11,
	"parameter-problem":       12,
	"timestamp-request":       13,
	"timestamp-reply":         14,
	"information-request":     15,
	"information-reply":       16,
	"addressmask-request":     17,
	"addressmask-reply":       18,
}

// icmpTypeReverseMap ICMP 타입 숫자 → 이름 매핑 (smartfw 동일)
var icmpTypeReverseMap = map[int]string{
	0:  "echo-reply",
	3:  "destination-unreachable",
	4:  "source-quench",
	5:  "echo-redirect",
	8:  "echo-request",
	11: "time-exceeded",
	12: "parameter-problem",
	13: "timestamp-request",
	14: "timestamp-reply",
	15: "information-request",
	16: "information-reply",
	17: "addressmask-request",
	18: "addressmask-reply",
}

// ICMPTypeNameToNumber ICMP type 이름을 숫자로 변환
func ICMPTypeNameToNumber(name string) (int, error) {
	// 이름으로 찾기
	if num, ok := icmpTypeMap[name]; ok {
		return num, nil
	}

	// 숫자 문자열인 경우
	if num, err := strconv.Atoi(name); err == nil {
		return num, nil
	}

	return 0, fmt.Errorf("알 수 없는 ICMP 타입: %s", name)
}

// ICMPTypeNumberToName ICMP type 숫자를 이름으로 변환
func ICMPTypeNumberToName(num int) string {
	if name, ok := icmpTypeReverseMap[num]; ok {
		return name
	}
	return strconv.Itoa(num)
}

// GetICMPCodeOptions ICMP code 옵션 목록 (Type 3: destination-unreachable 전용)
func GetICMPCodeOptions() []string {
	return []string{
		"None",
		"net-unreachable (0)",
		"host-unreachable (1)",
		"protocol-unreachable (2)",
		"port-unreachable (3)",
		"fragmentation-needed (4)",
		"source-route-failed (5)",
		"Custom...",
	}
}

// icmpCodeMap ICMP Code 이름 → 숫자 매핑 (Type 3: destination-unreachable)
var icmpCodeMap = map[string]int{
	"net-unreachable":       0,
	"host-unreachable":      1,
	"protocol-unreachable":  2,
	"port-unreachable":      3,
	"fragmentation-needed":  4,
	"source-route-failed":   5,
}

// icmpCodeReverseMap ICMP Code 숫자 → 이름 매핑
var icmpCodeReverseMap = map[int]string{
	0: "net-unreachable",
	1: "host-unreachable",
	2: "protocol-unreachable",
	3: "port-unreachable",
	4: "fragmentation-needed",
	5: "source-route-failed",
}

// ICMPCodeNameToNumber ICMP code 이름을 숫자로 변환
func ICMPCodeNameToNumber(name string) (int, error) {
	// 이름으로 찾기
	if num, ok := icmpCodeMap[name]; ok {
		return num, nil
	}

	// 숫자 문자열인 경우
	if num, err := strconv.Atoi(name); err == nil {
		return num, nil
	}

	return 0, fmt.Errorf("알 수 없는 ICMP 코드: %s", name)
}

// ICMPCodeNumberToName ICMP code 숫자를 이름으로 변환
func ICMPCodeNumberToName(num int) string {
	if name, ok := icmpCodeReverseMap[num]; ok {
		return name
	}
	return strconv.Itoa(num)
}
