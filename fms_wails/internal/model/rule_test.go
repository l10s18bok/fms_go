package model

import (
	"testing"
)

// TestChainConversion Chain 문자열 변환 테스트
func TestChainConversion(t *testing.T) {
	tests := []struct {
		chain    Chain
		expected string
	}{
		{ChainINPUT, "INPUT"},
		{ChainOUTPUT, "OUTPUT"},
		{ChainFORWARD, "FORWARD"},
		{ChainPREROUTING, "PREROUTING"},
		{ChainPOSTROUTING, "POSTROUTING"},
	}

	for _, tt := range tests {
		result := ChainToString(tt.chain)
		if result != tt.expected {
			t.Errorf("ChainToString(%d) = %s, want %s", tt.chain, result, tt.expected)
		}

		back := StringToChain(result)
		if back != tt.chain {
			t.Errorf("StringToChain(%s) = %d, want %d", result, back, tt.chain)
		}
	}
}

// TestStringToChain 문자열 → Chain 변환 테스트
func TestStringToChain(t *testing.T) {
	tests := []struct {
		input    string
		expected Chain
	}{
		{"INPUT", ChainINPUT},
		{"input", ChainINPUT},
		{"OUTPUT", ChainOUTPUT},
		{"output", ChainOUTPUT},
		{"FORWARD", ChainFORWARD},
		{"UNKNOWN", ChainINPUT}, // 기본값
	}

	for _, tt := range tests {
		result := StringToChain(tt.input)
		if result != tt.expected {
			t.Errorf("StringToChain(%s) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

// TestProtocolConversion Protocol 문자열 변환 테스트
func TestProtocolConversion(t *testing.T) {
	tests := []struct {
		protocol Protocol
		expected string
	}{
		{ProtocolTCP, "tcp"},
		{ProtocolUDP, "udp"},
		{ProtocolICMP, "icmp"},
		{ProtocolANY, "any"},
	}

	for _, tt := range tests {
		result := ProtocolToString(tt.protocol)
		if result != tt.expected {
			t.Errorf("ProtocolToString(%d) = %s, want %s", tt.protocol, result, tt.expected)
		}

		back := StringToProtocol(result)
		if back != tt.protocol {
			t.Errorf("StringToProtocol(%s) = %d, want %d", result, back, tt.protocol)
		}
	}
}

// TestStringToProtocol 문자열 → Protocol 변환 테스트
func TestStringToProtocol(t *testing.T) {
	tests := []struct {
		input    string
		expected Protocol
	}{
		{"tcp", ProtocolTCP},
		{"TCP", ProtocolTCP},
		{"udp", ProtocolUDP},
		{"UDP", ProtocolUDP},
		{"icmp", ProtocolICMP},
		{"any", ProtocolANY},
		{"unknown", ProtocolTCP}, // 기본값
	}

	for _, tt := range tests {
		result := StringToProtocol(tt.input)
		if result != tt.expected {
			t.Errorf("StringToProtocol(%s) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

// TestActionConversion Action 문자열 변환 테스트
func TestActionConversion(t *testing.T) {
	tests := []struct {
		action   Action
		expected string
	}{
		{ActionDROP, "DROP"},
		{ActionACCEPT, "ACCEPT"},
		{ActionREJECT, "REJECT"},
	}

	for _, tt := range tests {
		result := ActionToString(tt.action)
		if result != tt.expected {
			t.Errorf("ActionToString(%d) = %s, want %s", tt.action, result, tt.expected)
		}

		back := StringToAction(result)
		if back != tt.action {
			t.Errorf("StringToAction(%s) = %d, want %d", result, back, tt.action)
		}
	}
}

// TestGetChainOptions Chain 옵션 목록 테스트
func TestGetChainOptions(t *testing.T) {
	options := GetChainOptions()
	if len(options) != 3 {
		t.Errorf("GetChainOptions() length = %d, want 3", len(options))
	}
	if options[0] != "INPUT" {
		t.Errorf("GetChainOptions()[0] = %s, want INPUT", options[0])
	}
}

// TestGetProtocolOptions Protocol 옵션 목록 테스트
func TestGetProtocolOptions(t *testing.T) {
	options := GetProtocolOptions()
	if len(options) != 4 {
		t.Errorf("GetProtocolOptions() length = %d, want 4", len(options))
	}
	if options[0] != "tcp" {
		t.Errorf("GetProtocolOptions()[0] = %s, want tcp", options[0])
	}
}

// TestGetActionOptions Action 옵션 목록 테스트
func TestGetActionOptions(t *testing.T) {
	options := GetActionOptions()
	if len(options) != 2 {
		t.Errorf("GetActionOptions() length = %d, want 2", len(options))
	}
	if options[0] != "DROP" {
		t.Errorf("GetActionOptions()[0] = %s, want DROP", options[0])
	}
}

// TestNewFirewallRule 새 규칙 생성 테스트
func TestNewFirewallRule(t *testing.T) {
	rule := NewFirewallRule()

	if rule.Chain != ChainINPUT {
		t.Errorf("NewFirewallRule().Chain = %d, want %d", rule.Chain, ChainINPUT)
	}
	if rule.Protocol != ProtocolTCP {
		t.Errorf("NewFirewallRule().Protocol = %d, want %d", rule.Protocol, ProtocolTCP)
	}
	if rule.Action != ActionDROP {
		t.Errorf("NewFirewallRule().Action = %d, want %d", rule.Action, ActionDROP)
	}
}

// TestProtocolOptionsIsEmpty ProtocolOptions.IsEmpty() 테스트
func TestProtocolOptionsIsEmpty(t *testing.T) {
	// nil 체크
	var nilOpts *ProtocolOptions
	if !nilOpts.IsEmpty() {
		t.Error("nil ProtocolOptions.IsEmpty() should be true")
	}

	// 빈 옵션
	emptyOpts := &ProtocolOptions{}
	if !emptyOpts.IsEmpty() {
		t.Error("empty ProtocolOptions.IsEmpty() should be true")
	}

	// TCP Flags 있음
	tcpOpts := &ProtocolOptions{TCPFlags: "syn/syn"}
	if tcpOpts.IsEmpty() {
		t.Error("ProtocolOptions with TCPFlags.IsEmpty() should be false")
	}

	// ICMP Type 있음
	icmpOpts := &ProtocolOptions{ICMPType: "echo-request"}
	if icmpOpts.IsEmpty() {
		t.Error("ProtocolOptions with ICMPType.IsEmpty() should be false")
	}
}

// TestProtocolOptionsHasTCPOptions ProtocolOptions.HasTCPOptions() 테스트
func TestProtocolOptionsHasTCPOptions(t *testing.T) {
	var nilOpts *ProtocolOptions
	if nilOpts.HasTCPOptions() {
		t.Error("nil ProtocolOptions.HasTCPOptions() should be false")
	}

	emptyOpts := &ProtocolOptions{}
	if emptyOpts.HasTCPOptions() {
		t.Error("empty ProtocolOptions.HasTCPOptions() should be false")
	}

	tcpOpts := &ProtocolOptions{TCPFlags: "syn/syn"}
	if !tcpOpts.HasTCPOptions() {
		t.Error("ProtocolOptions with TCPFlags.HasTCPOptions() should be true")
	}
}

// TestProtocolOptionsHasICMPOptions ProtocolOptions.HasICMPOptions() 테스트
func TestProtocolOptionsHasICMPOptions(t *testing.T) {
	var nilOpts *ProtocolOptions
	if nilOpts.HasICMPOptions() {
		t.Error("nil ProtocolOptions.HasICMPOptions() should be false")
	}

	emptyOpts := &ProtocolOptions{}
	if emptyOpts.HasICMPOptions() {
		t.Error("empty ProtocolOptions.HasICMPOptions() should be false")
	}

	icmpTypeOpts := &ProtocolOptions{ICMPType: "echo-request"}
	if !icmpTypeOpts.HasICMPOptions() {
		t.Error("ProtocolOptions with ICMPType.HasICMPOptions() should be true")
	}

	icmpCodeOpts := &ProtocolOptions{ICMPCode: "3"}
	if !icmpCodeOpts.HasICMPOptions() {
		t.Error("ProtocolOptions with ICMPCode.HasICMPOptions() should be true")
	}
}

// TestTCPFlagsPresetToFlagsString TCPFlagsPreset.ToFlagsString() 테스트
func TestTCPFlagsPresetToFlagsString(t *testing.T) {
	tests := []struct {
		name     string
		preset   TCPFlagsPreset
		expected string
	}{
		{
			name: "None preset",
			preset: TCPFlagsPreset{
				Name:      "None",
				MaskFlags: nil,
				SetFlags:  nil,
			},
			expected: "",
		},
		{
			name: "New Connection (SYN)",
			preset: TCPFlagsPreset{
				Name:      "New Connection (SYN)",
				MaskFlags: []string{"syn", "rst", "ack", "fin"},
				SetFlags:  []string{"syn"},
			},
			expected: "syn,rst,ack,fin/syn",
		},
		{
			name: "Block NULL Scan",
			preset: TCPFlagsPreset{
				Name:      "Block NULL Scan",
				MaskFlags: []string{"syn", "rst", "ack", "fin", "psh", "urg"},
				SetFlags:  nil,
			},
			expected: "syn,rst,ack,fin,psh,urg/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.preset.ToFlagsString()
			if result != tt.expected {
				t.Errorf("ToFlagsString() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestGetTCPFlagsPresets 프리셋 목록 테스트
func TestGetTCPFlagsPresets(t *testing.T) {
	presets := GetTCPFlagsPresets()

	// 최소 2개 이상의 프리셋이 있어야 함
	if len(presets) < 2 {
		t.Errorf("GetTCPFlagsPresets() length = %d, want >= 2", len(presets))
	}

	// 첫 번째는 "None"
	if presets[0].Name != "None" {
		t.Errorf("GetTCPFlagsPresets()[0].Name = %s, want None", presets[0].Name)
	}

	// 마지막은 "Custom"
	last := presets[len(presets)-1]
	if last.Name != "Custom" {
		t.Errorf("GetTCPFlagsPresets() last.Name = %s, want Custom", last.Name)
	}
}

// TestFindPresetByFlags FindPresetByFlags() 테스트
func TestFindPresetByFlags(t *testing.T) {
	// 빈 문자열 → None
	preset := FindPresetByFlags("")
	if preset.Name != "None" {
		t.Errorf("FindPresetByFlags('') = %s, want None", preset.Name)
	}

	// SYN 프리셋
	synPreset := FindPresetByFlags("syn,rst,ack,fin/syn")
	if synPreset.Name != "New Connection (SYN)" {
		t.Errorf("FindPresetByFlags('syn,rst,ack,fin/syn') = %s, want New Connection (SYN)", synPreset.Name)
	}

	// 알 수 없는 플래그 → Custom
	customPreset := FindPresetByFlags("unknown/flags")
	if customPreset.Name != "Custom" {
		t.Errorf("FindPresetByFlags('unknown/flags') = %s, want Custom", customPreset.Name)
	}
}

// TestGetTCPFlagsList TCP flags 목록 테스트
func TestGetTCPFlagsList(t *testing.T) {
	flags := GetTCPFlagsList()
	expected := []string{"syn", "ack", "fin", "rst", "psh", "urg"}

	if len(flags) != len(expected) {
		t.Errorf("GetTCPFlagsList() length = %d, want %d", len(flags), len(expected))
	}

	for i, flag := range expected {
		if flags[i] != flag {
			t.Errorf("GetTCPFlagsList()[%d] = %s, want %s", i, flags[i], flag)
		}
	}
}

// TestICMPTypeConversion ICMP Type 변환 테스트
func TestICMPTypeConversion(t *testing.T) {
	tests := []struct {
		name   string
		number int
	}{
		{"echo-reply", 0},
		{"destination-unreachable", 3},
		{"echo-request", 8},
		{"time-exceeded", 11},
	}

	for _, tt := range tests {
		// 이름 → 숫자
		num, err := ICMPTypeNameToNumber(tt.name)
		if err != nil {
			t.Errorf("ICMPTypeNameToNumber(%s) error: %v", tt.name, err)
		}
		if num != tt.number {
			t.Errorf("ICMPTypeNameToNumber(%s) = %d, want %d", tt.name, num, tt.number)
		}

		// 숫자 → 이름
		name := ICMPTypeNumberToName(tt.number)
		if name != tt.name {
			t.Errorf("ICMPTypeNumberToName(%d) = %s, want %s", tt.number, name, tt.name)
		}
	}

	// 숫자 문자열 테스트
	num, err := ICMPTypeNameToNumber("8")
	if err != nil || num != 8 {
		t.Errorf("ICMPTypeNameToNumber('8') = %d, %v, want 8, nil", num, err)
	}

	// 알 수 없는 이름 테스트
	_, err = ICMPTypeNameToNumber("unknown")
	if err == nil {
		t.Error("ICMPTypeNameToNumber('unknown') should return error")
	}
}

// TestICMPCodeConversion ICMP Code 변환 테스트
func TestICMPCodeConversion(t *testing.T) {
	tests := []struct {
		name   string
		number int
	}{
		{"net-unreachable", 0},
		{"host-unreachable", 1},
		{"port-unreachable", 3},
	}

	for _, tt := range tests {
		// 이름 → 숫자
		num, err := ICMPCodeNameToNumber(tt.name)
		if err != nil {
			t.Errorf("ICMPCodeNameToNumber(%s) error: %v", tt.name, err)
		}
		if num != tt.number {
			t.Errorf("ICMPCodeNameToNumber(%s) = %d, want %d", tt.name, num, tt.number)
		}

		// 숫자 → 이름
		name := ICMPCodeNumberToName(tt.number)
		if name != tt.name {
			t.Errorf("ICMPCodeNumberToName(%d) = %s, want %s", tt.number, name, tt.name)
		}
	}
}

// TestGetICMPTypeOptions ICMP Type 옵션 목록 테스트
func TestGetICMPTypeOptions(t *testing.T) {
	options := GetICMPTypeOptions()

	// 최소 5개 이상
	if len(options) < 5 {
		t.Errorf("GetICMPTypeOptions() length = %d, want >= 5", len(options))
	}

	// 첫 번째는 "None"
	if options[0] != "None" {
		t.Errorf("GetICMPTypeOptions()[0] = %s, want None", options[0])
	}
}

// TestGetICMPCodeOptions ICMP Code 옵션 목록 테스트
func TestGetICMPCodeOptions(t *testing.T) {
	options := GetICMPCodeOptions()

	// 최소 3개 이상
	if len(options) < 3 {
		t.Errorf("GetICMPCodeOptions() length = %d, want >= 3", len(options))
	}

	// 첫 번째는 "None"
	if options[0] != "None" {
		t.Errorf("GetICMPCodeOptions()[0] = %s, want None", options[0])
	}
}
