package model_test

import (
	"testing"

	"fms/internal/model"
)

func TestGetTCPFlagsPresets(t *testing.T) {
	presets := model.GetTCPFlagsPresets()

	// 최소 7개의 프리셋이 있어야 함
	if len(presets) < 7 {
		t.Errorf("GetTCPFlagsPresets() returned %d presets, want at least 7", len(presets))
	}

	// 첫 번째는 "없음"
	if presets[0].Name != "없음" {
		t.Errorf("First preset should be '없음', got '%s'", presets[0].Name)
	}

	// 마지막은 "커스텀"
	if presets[len(presets)-1].Name != "커스텀" {
		t.Errorf("Last preset should be '커스텀', got '%s'", presets[len(presets)-1].Name)
	}
}

func TestTCPFlagsPreset_ToFlagsString(t *testing.T) {
	tests := []struct {
		name     string
		preset   model.TCPFlagsPreset
		expected string
	}{
		{
			name: "없음 (빈 문자열)",
			preset: model.TCPFlagsPreset{
				Name:      "없음",
				MaskFlags: nil,
				SetFlags:  nil,
			},
			expected: "",
		},
		{
			name: "새 연결만 (SYN)",
			preset: model.TCPFlagsPreset{
				Name:      "새 연결만 (SYN)",
				MaskFlags: []string{"syn", "rst", "ack", "fin"},
				SetFlags:  []string{"syn"},
			},
			expected: "syn,rst,ack,fin/syn",
		},
		{
			name: "확립된 연결 (ACK)",
			preset: model.TCPFlagsPreset{
				Name:      "확립된 연결 (ACK)",
				MaskFlags: []string{"ack"},
				SetFlags:  []string{"ack"},
			},
			expected: "ack/ack",
		},
		{
			name: "NULL 스캔 차단 (빈 SetFlags)",
			preset: model.TCPFlagsPreset{
				Name:      "NULL 스캔 차단",
				MaskFlags: []string{"syn", "rst", "ack", "fin", "psh", "urg"},
				SetFlags:  nil,
			},
			expected: "syn,rst,ack,fin,psh,urg/",
		},
		{
			name: "XMAS 스캔 차단",
			preset: model.TCPFlagsPreset{
				Name:      "XMAS 스캔 차단",
				MaskFlags: []string{"syn", "rst", "ack", "fin", "psh", "urg"},
				SetFlags:  []string{"fin", "psh", "urg"},
			},
			expected: "syn,rst,ack,fin,psh,urg/fin,psh,urg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.preset.ToFlagsString()
			if result != tt.expected {
				t.Errorf("ToFlagsString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFindPresetByFlags(t *testing.T) {
	tests := []struct {
		name         string
		flags        string
		expectedName string
	}{
		{
			name:         "빈 문자열 -> 없음",
			flags:        "",
			expectedName: "없음",
		},
		{
			name:         "새 연결만 (SYN)",
			flags:        "syn,rst,ack,fin/syn",
			expectedName: "새 연결만 (SYN)",
		},
		{
			name:         "확립된 연결 (ACK)",
			flags:        "ack/ack",
			expectedName: "확립된 연결 (ACK)",
		},
		{
			name:         "NULL 스캔 차단",
			flags:        "syn,rst,ack,fin,psh,urg/",
			expectedName: "NULL 스캔 차단",
		},
		{
			name:         "알 수 없는 패턴 -> 커스텀",
			flags:        "syn,ack/syn,ack",
			expectedName: "커스텀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset := model.FindPresetByFlags(tt.flags)
			if preset == nil {
				t.Fatalf("FindPresetByFlags(%q) returned nil", tt.flags)
			}
			if preset.Name != tt.expectedName {
				t.Errorf("FindPresetByFlags(%q).Name = %q, want %q", tt.flags, preset.Name, tt.expectedName)
			}
		})
	}
}

func TestGetTCPFlagsList(t *testing.T) {
	flags := model.GetTCPFlagsList()

	expected := []string{"syn", "ack", "fin", "rst", "psh", "urg"}

	if len(flags) != len(expected) {
		t.Errorf("GetTCPFlagsList() returned %d flags, want %d", len(flags), len(expected))
	}

	for i, flag := range expected {
		if flags[i] != flag {
			t.Errorf("GetTCPFlagsList()[%d] = %q, want %q", i, flags[i], flag)
		}
	}
}

func TestGetICMPTypeOptions(t *testing.T) {
	options := model.GetICMPTypeOptions()

	// 최소 7개의 옵션이 있어야 함
	if len(options) < 7 {
		t.Errorf("GetICMPTypeOptions() returned %d options, want at least 7", len(options))
	}

	// 첫 번째는 "없음"
	if options[0] != "없음" {
		t.Errorf("First option should be '없음', got '%s'", options[0])
	}
}

func TestICMPTypeNameToNumber(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		expected    int
		expectError bool
	}{
		{"echo-reply", "echo-reply", 0, false},
		{"destination-unreachable", "destination-unreachable", 3, false},
		{"source-quench", "source-quench", 4, false},
		{"redirect", "redirect", 5, false},
		{"echo-request", "echo-request", 8, false},
		{"time-exceeded", "time-exceeded", 11, false},
		{"parameter-problem", "parameter-problem", 12, false},
		{"timestamp-request", "timestamp-request", 13, false},
		{"timestamp-reply", "timestamp-reply", 14, false},
		{"숫자 문자열", "8", 8, false},
		{"알 수 없는 이름", "unknown", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := model.ICMPTypeNameToNumber(tt.typeName)
			if tt.expectError {
				if err == nil {
					t.Errorf("ICMPTypeNameToNumber(%q) expected error, got nil", tt.typeName)
				}
			} else {
				if err != nil {
					t.Errorf("ICMPTypeNameToNumber(%q) unexpected error: %v", tt.typeName, err)
				}
				if result != tt.expected {
					t.Errorf("ICMPTypeNameToNumber(%q) = %d, want %d", tt.typeName, result, tt.expected)
				}
			}
		})
	}
}

func TestICMPTypeNumberToName(t *testing.T) {
	tests := []struct {
		num      int
		expected string
	}{
		{0, "echo-reply"},
		{3, "destination-unreachable"},
		{4, "source-quench"},
		{5, "redirect"},
		{8, "echo-request"},
		{11, "time-exceeded"},
		{12, "parameter-problem"},
		{13, "timestamp-request"},
		{14, "timestamp-reply"},
		{99, "99"}, // 알 수 없는 숫자는 문자열로 반환
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := model.ICMPTypeNumberToName(tt.num)
			if result != tt.expected {
				t.Errorf("ICMPTypeNumberToName(%d) = %q, want %q", tt.num, result, tt.expected)
			}
		})
	}
}
