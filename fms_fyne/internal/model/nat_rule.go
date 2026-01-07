package model

import "strings"

// NATType NAT 규칙 타입
type NATType int

const (
	NATTypeDNAT       NATType = 0 // Destination NAT (포트 포워딩)
	NATTypeSNAT       NATType = 1 // Source NAT
	NATTypeMASQUERADE NATType = 2 // Masquerade
)

// NATRule NAT 규칙 구조체
type NATRule struct {
	// 기본 필드
	NATType  NATType  // DNAT, SNAT, MASQUERADE
	Protocol Protocol // TCP, UDP, ANY

	// 매칭 조건
	MatchIP   string // 매칭할 IP (소스 또는 목적지)
	MatchPort string // 매칭할 포트

	// 변환 대상
	TranslateIP   string // 변환할 IP
	TranslatePort string // 변환할 포트

	// 인터페이스
	InInterface  string // 입력 인터페이스 (예: eth0)
	OutInterface string // 출력 인터페이스 (예: eth1)

	// 추가 옵션
	Description string // 규칙 설명 (선택)
}

// NewNATRule 기본값으로 새 NAT 규칙 생성
func NewNATRule() *NATRule {
	return &NATRule{
		NATType:  NATTypeDNAT,
		Protocol: ProtocolTCP,
		MatchIP:  "ANY",
	}
}

// NewDNATRule DNAT 규칙 생성
func NewDNATRule() *NATRule {
	return &NATRule{
		NATType:  NATTypeDNAT,
		Protocol: ProtocolTCP,
		MatchIP:  "ANY",
	}
}

// NewSNATRule SNAT 규칙 생성
func NewSNATRule() *NATRule {
	return &NATRule{
		NATType:  NATTypeSNAT,
		Protocol: ProtocolTCP,
	}
}

// NewMASQUERADERule MASQUERADE 규칙 생성
func NewMASQUERADERule() *NATRule {
	return &NATRule{
		NATType:  NATTypeMASQUERADE,
		Protocol: ProtocolTCP,
	}
}

// NATTypeToString NATType을 문자열로 변환
func NATTypeToString(t NATType) string {
	switch t {
	case NATTypeDNAT:
		return "DNAT"
	case NATTypeSNAT:
		return "SNAT"
	case NATTypeMASQUERADE:
		return "MASQUERADE"
	default:
		return "DNAT"
	}
}

// StringToNATType 문자열을 NATType으로 변환
func StringToNATType(s string) NATType {
	switch strings.ToUpper(s) {
	case "DNAT":
		return NATTypeDNAT
	case "SNAT":
		return NATTypeSNAT
	case "MASQUERADE", "MASQ":
		return NATTypeMASQUERADE
	default:
		return NATTypeDNAT
	}
}

// GetNATTypeOptions UI Select용 NAT 타입 옵션 목록
func GetNATTypeOptions() []string {
	return []string{"DNAT", "SNAT", "MASQUERADE"}
}

// GetSNATTypeOptions SNAT 폼용 타입 옵션 (SNAT, MASQUERADE만)
func GetSNATTypeOptions() []string {
	return []string{"SNAT", "MASQUERADE"}
}

// NATTypeToDisplayString UI 표시용 문자열
func NATTypeToDisplayString(t NATType) string {
	switch t {
	case NATTypeDNAT:
		return "DNAT (포트 포워딩)"
	case NATTypeSNAT:
		return "SNAT (소스 NAT)"
	case NATTypeMASQUERADE:
		return "MASQUERADE"
	default:
		return "DNAT (포트 포워딩)"
	}
}
