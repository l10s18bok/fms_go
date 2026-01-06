package model_test

import (
	"testing"

	"fms/internal/model"
)

func TestProtocolOptions_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		opts     *model.ProtocolOptions
		expected bool
	}{
		{
			name:     "nil options",
			opts:     nil,
			expected: true,
		},
		{
			name:     "empty options",
			opts:     &model.ProtocolOptions{},
			expected: true,
		},
		{
			name:     "with TCP flags",
			opts:     &model.ProtocolOptions{TCPFlags: "syn/syn"},
			expected: false,
		},
		{
			name:     "with ICMP type",
			opts:     &model.ProtocolOptions{ICMPType: "echo-request"},
			expected: false,
		},
		{
			name:     "with ICMP code only",
			opts:     &model.ProtocolOptions{ICMPCode: "0"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.IsEmpty()
			if result != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProtocolOptions_HasTCPOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     *model.ProtocolOptions
		expected bool
	}{
		{
			name:     "nil options",
			opts:     nil,
			expected: false,
		},
		{
			name:     "empty options",
			opts:     &model.ProtocolOptions{},
			expected: false,
		},
		{
			name:     "with TCP flags",
			opts:     &model.ProtocolOptions{TCPFlags: "syn/syn"},
			expected: true,
		},
		{
			name:     "with ICMP type only",
			opts:     &model.ProtocolOptions{ICMPType: "echo-request"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.HasTCPOptions()
			if result != tt.expected {
				t.Errorf("HasTCPOptions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProtocolOptions_HasICMPOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     *model.ProtocolOptions
		expected bool
	}{
		{
			name:     "nil options",
			opts:     nil,
			expected: false,
		},
		{
			name:     "empty options",
			opts:     &model.ProtocolOptions{},
			expected: false,
		},
		{
			name:     "with ICMP type",
			opts:     &model.ProtocolOptions{ICMPType: "echo-request"},
			expected: true,
		},
		{
			name:     "with ICMP code",
			opts:     &model.ProtocolOptions{ICMPCode: "0"},
			expected: true,
		},
		{
			name:     "with TCP flags only",
			opts:     &model.ProtocolOptions{TCPFlags: "syn/syn"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.HasICMPOptions()
			if result != tt.expected {
				t.Errorf("HasICMPOptions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetICMPCodeOptions(t *testing.T) {
	options := model.GetICMPCodeOptions()

	// 최소 필수 옵션 확인
	if len(options) < 7 {
		t.Errorf("GetICMPCodeOptions() 옵션 개수가 부족합니다: %d", len(options))
	}

	// 첫번째 옵션은 "없음"
	if options[0] != "없음" {
		t.Errorf("첫번째 옵션이 '없음'이 아닙니다: %s", options[0])
	}

	// 마지막 옵션은 "커스텀 숫자..."
	if options[len(options)-1] != "커스텀 숫자..." {
		t.Errorf("마지막 옵션이 '커스텀 숫자...'가 아닙니다: %s", options[len(options)-1])
	}
}

func TestICMPCodeNameToNumber(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{
			name:        "net-unreachable",
			input:       "net-unreachable",
			expected:    0,
			expectError: false,
		},
		{
			name:        "host-unreachable",
			input:       "host-unreachable",
			expected:    1,
			expectError: false,
		},
		{
			name:        "protocol-unreachable",
			input:       "protocol-unreachable",
			expected:    2,
			expectError: false,
		},
		{
			name:        "port-unreachable",
			input:       "port-unreachable",
			expected:    3,
			expectError: false,
		},
		{
			name:        "fragmentation-needed",
			input:       "fragmentation-needed",
			expected:    4,
			expectError: false,
		},
		{
			name:        "source-route-failed",
			input:       "source-route-failed",
			expected:    5,
			expectError: false,
		},
		{
			name:        "숫자 문자열",
			input:       "3",
			expected:    3,
			expectError: false,
		},
		{
			name:        "알 수 없는 이름",
			input:       "unknown",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := model.ICMPCodeNameToNumber(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("에러가 발생해야 합니다")
				}
			} else {
				if err != nil {
					t.Errorf("예상치 못한 에러: %v", err)
				}
				if result != tt.expected {
					t.Errorf("ICMPCodeNameToNumber(%s) = %d, want %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestICMPCodeNumberToName(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "code 0",
			input:    0,
			expected: "net-unreachable",
		},
		{
			name:     "code 1",
			input:    1,
			expected: "host-unreachable",
		},
		{
			name:     "code 2",
			input:    2,
			expected: "protocol-unreachable",
		},
		{
			name:     "code 3",
			input:    3,
			expected: "port-unreachable",
		},
		{
			name:     "code 4",
			input:    4,
			expected: "fragmentation-needed",
		},
		{
			name:     "code 5",
			input:    5,
			expected: "source-route-failed",
		},
		{
			name:     "알 수 없는 숫자",
			input:    99,
			expected: "99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ICMPCodeNumberToName(tt.input)
			if result != tt.expected {
				t.Errorf("ICMPCodeNumberToName(%d) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}
