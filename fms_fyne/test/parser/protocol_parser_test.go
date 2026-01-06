package parser_test

import (
	"testing"

	"fms/internal/model"
	"fms/internal/parser"
)

func TestParseProtocolWithOptions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantProto    model.Protocol
		wantTCPFlags string
		wantICMPType string
		wantICMPCode string
		wantErr      bool
	}{
		{
			name:      "기본 TCP (옵션 없음)",
			input:     "tcp",
			wantProto: model.ProtocolTCP,
		},
		{
			name:      "기본 UDP",
			input:     "udp",
			wantProto: model.ProtocolUDP,
		},
		{
			name:      "기본 ICMP",
			input:     "icmp",
			wantProto: model.ProtocolICMP,
		},
		{
			name:      "기본 ANY",
			input:     "any",
			wantProto: model.ProtocolANY,
		},
		{
			name:         "TCP flags 단일",
			input:        "tcp?flags=syn/syn",
			wantProto:    model.ProtocolTCP,
			wantTCPFlags: "syn/syn",
		},
		{
			name:         "TCP flags 복수",
			input:        "tcp?flags=syn,ack/syn",
			wantProto:    model.ProtocolTCP,
			wantTCPFlags: "syn,ack/syn",
		},
		{
			name:         "TCP flags 새 연결",
			input:        "tcp?flags=syn,rst,ack,fin/syn",
			wantProto:    model.ProtocolTCP,
			wantTCPFlags: "syn,rst,ack,fin/syn",
		},
		{
			name:         "ICMP type 이름",
			input:        "icmp?type=echo-request",
			wantProto:    model.ProtocolICMP,
			wantICMPType: "echo-request",
		},
		{
			name:         "ICMP type 숫자",
			input:        "icmp?type=8",
			wantProto:    model.ProtocolICMP,
			wantICMPType: "8",
		},
		{
			name:         "ICMP type + code",
			input:        "icmp?type=3&code=0",
			wantProto:    model.ProtocolICMP,
			wantICMPType: "3",
			wantICMPCode: "0",
		},
		{
			name:         "ICMP type 이름 + code",
			input:        "icmp?type=destination-unreachable&code=3",
			wantProto:    model.ProtocolICMP,
			wantICMPType: "destination-unreachable",
			wantICMPCode: "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto, opts, err := parser.ParseProtocolWithOptions(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseProtocolWithOptions(%q) expected error", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseProtocolWithOptions(%q) unexpected error: %v", tt.input, err)
				return
			}

			if proto != tt.wantProto {
				t.Errorf("ParseProtocolWithOptions(%q) proto = %v, want %v", tt.input, proto, tt.wantProto)
			}

			// 옵션 검사
			if tt.wantTCPFlags != "" || tt.wantICMPType != "" || tt.wantICMPCode != "" {
				if opts == nil {
					t.Errorf("ParseProtocolWithOptions(%q) opts = nil, want non-nil", tt.input)
					return
				}

				if opts.TCPFlags != tt.wantTCPFlags {
					t.Errorf("ParseProtocolWithOptions(%q) TCPFlags = %q, want %q", tt.input, opts.TCPFlags, tt.wantTCPFlags)
				}
				if opts.ICMPType != tt.wantICMPType {
					t.Errorf("ParseProtocolWithOptions(%q) ICMPType = %q, want %q", tt.input, opts.ICMPType, tt.wantICMPType)
				}
				if opts.ICMPCode != tt.wantICMPCode {
					t.Errorf("ParseProtocolWithOptions(%q) ICMPCode = %q, want %q", tt.input, opts.ICMPCode, tt.wantICMPCode)
				}
			} else {
				// 옵션 없으면 nil이거나 비어있어야 함
				if opts != nil && !opts.IsEmpty() {
					t.Errorf("ParseProtocolWithOptions(%q) opts should be nil or empty", tt.input)
				}
			}
		})
	}
}

func TestFormatProtocolWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		proto    model.Protocol
		opts     *model.ProtocolOptions
		expected string
	}{
		{
			name:     "TCP 옵션 없음",
			proto:    model.ProtocolTCP,
			opts:     nil,
			expected: "tcp",
		},
		{
			name:     "UDP 옵션 없음",
			proto:    model.ProtocolUDP,
			opts:     nil,
			expected: "udp",
		},
		{
			name:     "ICMP 옵션 없음",
			proto:    model.ProtocolICMP,
			opts:     nil,
			expected: "icmp",
		},
		{
			name:     "ANY 옵션 없음",
			proto:    model.ProtocolANY,
			opts:     nil,
			expected: "any",
		},
		{
			name:     "TCP 빈 옵션",
			proto:    model.ProtocolTCP,
			opts:     &model.ProtocolOptions{},
			expected: "tcp",
		},
		{
			name:     "TCP flags",
			proto:    model.ProtocolTCP,
			opts:     &model.ProtocolOptions{TCPFlags: "syn/syn"},
			expected: "tcp?flags=syn/syn",
		},
		{
			name:     "TCP flags 복잡",
			proto:    model.ProtocolTCP,
			opts:     &model.ProtocolOptions{TCPFlags: "syn,rst,ack,fin/syn"},
			expected: "tcp?flags=syn,rst,ack,fin/syn",
		},
		{
			name:     "ICMP type만",
			proto:    model.ProtocolICMP,
			opts:     &model.ProtocolOptions{ICMPType: "echo-request"},
			expected: "icmp?type=echo-request",
		},
		{
			name:     "ICMP type + code",
			proto:    model.ProtocolICMP,
			opts:     &model.ProtocolOptions{ICMPType: "3", ICMPCode: "0"},
			expected: "icmp?type=3&code=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.FormatProtocolWithOptions(tt.proto, tt.opts)
			if result != tt.expected {
				t.Errorf("FormatProtocolWithOptions() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// 파싱 → 포맷 → 파싱 일관성 확인
	tests := []string{
		"tcp",
		"udp",
		"icmp",
		"any",
		"tcp?flags=syn/syn",
		"tcp?flags=syn,ack/syn",
		"tcp?flags=syn,rst,ack,fin/syn",
		"icmp?type=echo-request",
		"icmp?type=8",
		"icmp?type=3&code=0",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// 첫 번째 파싱
			proto1, opts1, err := parser.ParseProtocolWithOptions(input)
			if err != nil {
				t.Fatalf("First parse error: %v", err)
			}

			// 포맷
			formatted := parser.FormatProtocolWithOptions(proto1, opts1)

			// 두 번째 파싱
			proto2, opts2, err := parser.ParseProtocolWithOptions(formatted)
			if err != nil {
				t.Fatalf("Second parse error: %v", err)
			}

			// 비교
			if proto1 != proto2 {
				t.Errorf("Protocol mismatch: %v != %v", proto1, proto2)
			}

			// 옵션 비교
			if opts1 == nil && opts2 == nil {
				return
			}
			if opts1 == nil || opts2 == nil {
				// 하나만 nil인 경우 비어있는지 확인
				if opts1 != nil && !opts1.IsEmpty() {
					t.Errorf("Options mismatch: opts1 not empty, opts2 nil")
				}
				if opts2 != nil && !opts2.IsEmpty() {
					t.Errorf("Options mismatch: opts1 nil, opts2 not empty")
				}
				return
			}

			if opts1.TCPFlags != opts2.TCPFlags {
				t.Errorf("TCPFlags mismatch: %q != %q", opts1.TCPFlags, opts2.TCPFlags)
			}
			if opts1.ICMPType != opts2.ICMPType {
				t.Errorf("ICMPType mismatch: %q != %q", opts1.ICMPType, opts2.ICMPType)
			}
			if opts1.ICMPCode != opts2.ICMPCode {
				t.Errorf("ICMPCode mismatch: %q != %q", opts1.ICMPCode, opts2.ICMPCode)
			}
		})
	}
}

func TestParseLine_WithOptions(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantProto    model.Protocol
		wantTCPFlags string
		wantICMPType string
		wantICMPCode string
	}{
		{
			name:      "기존 형식 (옵션 없음)",
			line:      "agent -m=insert -c=INPUT -p=tcp -a=DROP",
			wantProto: model.ProtocolTCP,
		},
		{
			name:         "TCP flags",
			line:         "agent -m=insert -c=INPUT -p=tcp?flags=syn/syn -a=DROP --dport=80",
			wantProto:    model.ProtocolTCP,
			wantTCPFlags: "syn/syn",
		},
		{
			name:         "ICMP type",
			line:         "agent -m=insert -c=INPUT -p=icmp?type=echo-request -a=DROP",
			wantProto:    model.ProtocolICMP,
			wantICMPType: "echo-request",
		},
		{
			name:         "ICMP type + code",
			line:         "agent -m=insert -c=INPUT -p=icmp?type=3&code=0 -a=DROP",
			wantProto:    model.ProtocolICMP,
			wantICMPType: "3",
			wantICMPCode: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := parser.ParseLine(tt.line)
			if err != nil {
				t.Fatalf("ParseLine error: %v", err)
			}
			if rule == nil {
				t.Fatalf("ParseLine returned nil rule")
			}

			if rule.Protocol != tt.wantProto {
				t.Errorf("Protocol = %v, want %v", rule.Protocol, tt.wantProto)
			}

			if tt.wantTCPFlags != "" || tt.wantICMPType != "" {
				if rule.Options == nil {
					t.Fatalf("Options = nil, want non-nil")
				}
				if rule.Options.TCPFlags != tt.wantTCPFlags {
					t.Errorf("TCPFlags = %q, want %q", rule.Options.TCPFlags, tt.wantTCPFlags)
				}
				if rule.Options.ICMPType != tt.wantICMPType {
					t.Errorf("ICMPType = %q, want %q", rule.Options.ICMPType, tt.wantICMPType)
				}
				if rule.Options.ICMPCode != tt.wantICMPCode {
					t.Errorf("ICMPCode = %q, want %q", rule.Options.ICMPCode, tt.wantICMPCode)
				}
			}
		})
	}
}

func TestRuleToLine_WithOptions(t *testing.T) {
	tests := []struct {
		name     string
		rule     *model.FirewallRule
		contains []string // 결과에 포함되어야 하는 문자열들
	}{
		{
			name: "옵션 없음",
			rule: &model.FirewallRule{
				Chain:    model.ChainINPUT,
				Protocol: model.ProtocolTCP,
				Action:   model.ActionDROP,
			},
			contains: []string{"-p=tcp", "-a=DROP"},
		},
		{
			name: "TCP flags",
			rule: &model.FirewallRule{
				Chain:    model.ChainINPUT,
				Protocol: model.ProtocolTCP,
				Options:  &model.ProtocolOptions{TCPFlags: "syn/syn"},
				Action:   model.ActionDROP,
				DPort:    "80",
			},
			contains: []string{"-p=tcp?flags=syn/syn", "--dport=80"},
		},
		{
			name: "ICMP type",
			rule: &model.FirewallRule{
				Chain:    model.ChainINPUT,
				Protocol: model.ProtocolICMP,
				Options:  &model.ProtocolOptions{ICMPType: "echo-request"},
				Action:   model.ActionDROP,
			},
			contains: []string{"-p=icmp?type=echo-request"},
		},
		{
			name: "ICMP type + code",
			rule: &model.FirewallRule{
				Chain:    model.ChainINPUT,
				Protocol: model.ProtocolICMP,
				Options:  &model.ProtocolOptions{ICMPType: "3", ICMPCode: "0"},
				Action:   model.ActionDROP,
			},
			contains: []string{"-p=icmp?type=3&code=0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.RuleToLine(tt.rule)
			for _, s := range tt.contains {
				if !containsString(result, s) {
					t.Errorf("RuleToLine() = %q, want to contain %q", result, s)
				}
			}
		})
	}
}

func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || len(needle) == 0 ||
		(len(haystack) > len(needle) && findSubstring(haystack, needle)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
