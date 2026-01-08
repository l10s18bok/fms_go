package model

import "testing"

// TestNATTypeConversion NATType 문자열 변환 테스트
func TestNATTypeConversion(t *testing.T) {
	tests := []struct {
		natType  NATType
		expected string
	}{
		{NATTypeDNAT, "DNAT"},
		{NATTypeSNAT, "SNAT"},
		{NATTypeMASQUERADE, "MASQUERADE"},
	}

	for _, tt := range tests {
		result := NATTypeToString(tt.natType)
		if result != tt.expected {
			t.Errorf("NATTypeToString(%d) = %s, want %s", tt.natType, result, tt.expected)
		}

		back := StringToNATType(result)
		if back != tt.natType {
			t.Errorf("StringToNATType(%s) = %d, want %d", result, back, tt.natType)
		}
	}
}

// TestStringToNATType 문자열 → NATType 변환 테스트
func TestStringToNATType(t *testing.T) {
	tests := []struct {
		input    string
		expected NATType
	}{
		{"DNAT", NATTypeDNAT},
		{"dnat", NATTypeDNAT},
		{"SNAT", NATTypeSNAT},
		{"snat", NATTypeSNAT},
		{"MASQUERADE", NATTypeMASQUERADE},
		{"masquerade", NATTypeMASQUERADE},
		{"MASQ", NATTypeMASQUERADE},
		{"unknown", NATTypeDNAT}, // 기본값
	}

	for _, tt := range tests {
		result := StringToNATType(tt.input)
		if result != tt.expected {
			t.Errorf("StringToNATType(%s) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

// TestGetNATTypeOptions NAT 타입 옵션 목록 테스트
func TestGetNATTypeOptions(t *testing.T) {
	options := GetNATTypeOptions()
	if len(options) != 3 {
		t.Errorf("GetNATTypeOptions() length = %d, want 3", len(options))
	}
	if options[0] != "DNAT" {
		t.Errorf("GetNATTypeOptions()[0] = %s, want DNAT", options[0])
	}
}

// TestGetSNATTypeOptions SNAT 타입 옵션 목록 테스트
func TestGetSNATTypeOptions(t *testing.T) {
	options := GetSNATTypeOptions()
	if len(options) != 2 {
		t.Errorf("GetSNATTypeOptions() length = %d, want 2", len(options))
	}
	if options[0] != "SNAT" {
		t.Errorf("GetSNATTypeOptions()[0] = %s, want SNAT", options[0])
	}
}

// TestNewNATRule 새 NAT 규칙 생성 테스트
func TestNewNATRule(t *testing.T) {
	rule := NewNATRule()

	if rule.NATType != NATTypeDNAT {
		t.Errorf("NewNATRule().NATType = %d, want %d", rule.NATType, NATTypeDNAT)
	}
	if rule.Protocol != ProtocolTCP {
		t.Errorf("NewNATRule().Protocol = %d, want %d", rule.Protocol, ProtocolTCP)
	}
	if rule.MatchIP != "ANY" {
		t.Errorf("NewNATRule().MatchIP = %s, want ANY", rule.MatchIP)
	}
}

// TestNewDNATRule DNAT 규칙 생성 테스트
func TestNewDNATRule(t *testing.T) {
	rule := NewDNATRule()

	if rule.NATType != NATTypeDNAT {
		t.Errorf("NewDNATRule().NATType = %d, want %d", rule.NATType, NATTypeDNAT)
	}
}

// TestNewSNATRule SNAT 규칙 생성 테스트
func TestNewSNATRule(t *testing.T) {
	rule := NewSNATRule()

	if rule.NATType != NATTypeSNAT {
		t.Errorf("NewSNATRule().NATType = %d, want %d", rule.NATType, NATTypeSNAT)
	}
}

// TestNewMASQUERADERule MASQUERADE 규칙 생성 테스트
func TestNewMASQUERADERule(t *testing.T) {
	rule := NewMASQUERADERule()

	if rule.NATType != NATTypeMASQUERADE {
		t.Errorf("NewMASQUERADERule().NATType = %d, want %d", rule.NATType, NATTypeMASQUERADE)
	}
}

// TestNATTypeToDisplayString UI 표시용 문자열 테스트
func TestNATTypeToDisplayString(t *testing.T) {
	tests := []struct {
		natType  NATType
		expected string
	}{
		{NATTypeDNAT, "DNAT (포트 포워딩)"},
		{NATTypeSNAT, "SNAT (소스 NAT)"},
		{NATTypeMASQUERADE, "MASQUERADE"},
	}

	for _, tt := range tests {
		result := NATTypeToDisplayString(tt.natType)
		if result != tt.expected {
			t.Errorf("NATTypeToDisplayString(%d) = %s, want %s", tt.natType, result, tt.expected)
		}
	}
}
