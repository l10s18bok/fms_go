package parser

import (
	"testing"

	"fms_wails/internal/model"
)

// TestParseProtocolWithOptions 프로토콜 옵션 파싱 테스트
func TestParseProtocolWithOptions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantProtocol model.Protocol
		wantFlags    string
		wantType     string
		wantCode     string
	}{
		{
			name:         "TCP without options",
			input:        "tcp",
			wantProtocol: model.ProtocolTCP,
			wantFlags:    "",
			wantType:     "",
			wantCode:     "",
		},
		{
			name:         "TCP with flags",
			input:        "tcp?flags=syn/syn",
			wantProtocol: model.ProtocolTCP,
			wantFlags:    "syn/syn",
			wantType:     "",
			wantCode:     "",
		},
		{
			name:         "TCP with complex flags",
			input:        "tcp?flags=syn,rst,ack,fin/syn",
			wantProtocol: model.ProtocolTCP,
			wantFlags:    "syn,rst,ack,fin/syn",
			wantType:     "",
			wantCode:     "",
		},
		{
			name:         "ICMP with type",
			input:        "icmp?type=echo-request",
			wantProtocol: model.ProtocolICMP,
			wantFlags:    "",
			wantType:     "echo-request",
			wantCode:     "",
		},
		{
			name:         "ICMP with type and code",
			input:        "icmp?type=3&code=0",
			wantProtocol: model.ProtocolICMP,
			wantFlags:    "",
			wantType:     "3",
			wantCode:     "0",
		},
		{
			name:         "UDP without options",
			input:        "udp",
			wantProtocol: model.ProtocolUDP,
			wantFlags:    "",
			wantType:     "",
			wantCode:     "",
		},
		{
			name:         "ANY protocol",
			input:        "any",
			wantProtocol: model.ProtocolANY,
			wantFlags:    "",
			wantType:     "",
			wantCode:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto, opts, err := ParseProtocolWithOptions(tt.input)
			if err != nil {
				t.Errorf("ParseProtocolWithOptions() error = %v", err)
				return
			}

			if proto != tt.wantProtocol {
				t.Errorf("ParseProtocolWithOptions() protocol = %v, want %v", proto, tt.wantProtocol)
			}

			if tt.wantFlags == "" && tt.wantType == "" && tt.wantCode == "" {
				if opts != nil && !opts.IsEmpty() {
					t.Errorf("ParseProtocolWithOptions() opts should be nil or empty")
				}
				return
			}

			if opts == nil {
				t.Errorf("ParseProtocolWithOptions() opts should not be nil")
				return
			}

			if opts.TCPFlags != tt.wantFlags {
				t.Errorf("ParseProtocolWithOptions() flags = %v, want %v", opts.TCPFlags, tt.wantFlags)
			}
			if opts.ICMPType != tt.wantType {
				t.Errorf("ParseProtocolWithOptions() type = %v, want %v", opts.ICMPType, tt.wantType)
			}
			if opts.ICMPCode != tt.wantCode {
				t.Errorf("ParseProtocolWithOptions() code = %v, want %v", opts.ICMPCode, tt.wantCode)
			}
		})
	}
}

// TestFormatProtocolWithOptions 프로토콜 옵션 포맷 테스트
func TestFormatProtocolWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		protocol model.Protocol
		opts     *model.ProtocolOptions
		want     string
	}{
		{
			name:     "TCP without options",
			protocol: model.ProtocolTCP,
			opts:     nil,
			want:     "tcp",
		},
		{
			name:     "TCP with empty options",
			protocol: model.ProtocolTCP,
			opts:     &model.ProtocolOptions{},
			want:     "tcp",
		},
		{
			name:     "TCP with flags",
			protocol: model.ProtocolTCP,
			opts:     &model.ProtocolOptions{TCPFlags: "syn/syn"},
			want:     "tcp?flags=syn/syn",
		},
		{
			name:     "ICMP with type",
			protocol: model.ProtocolICMP,
			opts:     &model.ProtocolOptions{ICMPType: "echo-request"},
			want:     "icmp?type=echo-request",
		},
		{
			name:     "ICMP with type and code",
			protocol: model.ProtocolICMP,
			opts:     &model.ProtocolOptions{ICMPType: "3", ICMPCode: "0"},
			want:     "icmp?type=3&code=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatProtocolWithOptions(tt.protocol, tt.opts)
			if result != tt.want {
				t.Errorf("FormatProtocolWithOptions() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestFormatOptionsOnly 옵션만 포맷 테스트
func TestFormatOptionsOnly(t *testing.T) {
	tests := []struct {
		name string
		opts *model.ProtocolOptions
		want string
	}{
		{
			name: "nil options",
			opts: nil,
			want: "",
		},
		{
			name: "empty options",
			opts: &model.ProtocolOptions{},
			want: "",
		},
		{
			name: "TCP flags only",
			opts: &model.ProtocolOptions{TCPFlags: "syn/syn"},
			want: "flags=syn/syn",
		},
		{
			name: "ICMP type only",
			opts: &model.ProtocolOptions{ICMPType: "echo-request"},
			want: "type=echo-request",
		},
		{
			name: "ICMP type and code",
			opts: &model.ProtocolOptions{ICMPType: "3", ICMPCode: "0"},
			want: "type=3&code=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOptionsOnly(tt.opts)
			if result != tt.want {
				t.Errorf("FormatOptionsOnly() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestParseLine 단일 라인 파싱 테스트
func TestParseLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
		wantErr  bool
		validate func(*model.FirewallRule) bool
	}{
		{
			name:    "empty line",
			input:   "",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "comment line",
			input:   "# This is a comment",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid command",
			wantNil: true,
			wantErr: true,
		},
		{
			name:    "basic rule",
			input:   "agent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP",
			wantNil: false,
			wantErr: false,
			validate: func(r *model.FirewallRule) bool {
				return r.Chain == model.ChainINPUT &&
					r.Protocol == model.ProtocolTCP &&
					r.Action == model.ActionDROP &&
					r.DPort == "9010"
			},
		},
		{
			name:    "rule with SIP and DIP",
			input:   "agent -m=insert -c=OUTPUT -p=udp --sip=192.168.1.0/24 --dip=10.0.0.1 -a=ACCEPT",
			wantNil: false,
			wantErr: false,
			validate: func(r *model.FirewallRule) bool {
				return r.Chain == model.ChainOUTPUT &&
					r.Protocol == model.ProtocolUDP &&
					r.Action == model.ActionACCEPT &&
					r.SIP == "192.168.1.0/24" &&
					r.DIP == "10.0.0.1"
			},
		},
		{
			name:    "rule with black flag",
			input:   "agent -m=insert -c=INPUT -p=tcp -a=DROP --black",
			wantNil: false,
			wantErr: false,
			validate: func(r *model.FirewallRule) bool {
				return r.Black == true && r.White == false
			},
		},
		{
			name:    "rule with white flag",
			input:   "agent -m=insert -c=INPUT -p=tcp -a=ACCEPT --white",
			wantNil: false,
			wantErr: false,
			validate: func(r *model.FirewallRule) bool {
				return r.White == true && r.Black == false
			},
		},
		{
			name:    "rule with TCP flags option",
			input:   "agent -m=insert -c=INPUT -p=tcp?flags=syn/syn --dport=80 -a=DROP",
			wantNil: false,
			wantErr: false,
			validate: func(r *model.FirewallRule) bool {
				return r.Protocol == model.ProtocolTCP &&
					r.Options != nil &&
					r.Options.TCPFlags == "syn/syn" &&
					r.DPort == "80"
			},
		},
		{
			name:    "rule with ICMP type option",
			input:   "agent -m=insert -c=INPUT -p=icmp?type=echo-request -a=DROP",
			wantNil: false,
			wantErr: false,
			validate: func(r *model.FirewallRule) bool {
				return r.Protocol == model.ProtocolICMP &&
					r.Options != nil &&
					r.Options.ICMPType == "echo-request"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ParseLine(tt.input)

			if tt.wantErr && err == nil {
				t.Errorf("ParseLine() expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ParseLine() unexpected error: %v", err)
				return
			}

			if tt.wantNil && rule != nil {
				t.Errorf("ParseLine() expected nil rule, got %+v", rule)
				return
			}
			if !tt.wantNil && rule == nil {
				t.Errorf("ParseLine() expected rule, got nil")
				return
			}

			if tt.validate != nil && !tt.validate(rule) {
				t.Errorf("ParseLine() validation failed for rule: %+v", rule)
			}
		})
	}
}

// TestRuleToLine 규칙 → 텍스트 변환 테스트
func TestRuleToLine(t *testing.T) {
	tests := []struct {
		name string
		rule *model.FirewallRule
		want string
	}{
		{
			name: "nil rule",
			rule: nil,
			want: "",
		},
		{
			name: "basic rule",
			rule: &model.FirewallRule{
				Chain:    model.ChainINPUT,
				Protocol: model.ProtocolTCP,
				Action:   model.ActionDROP,
				DPort:    "9010",
			},
			want: "agent -m=insert -c=INPUT -p=tcp -a=DROP --dport=9010",
		},
		{
			name: "rule with all fields",
			rule: &model.FirewallRule{
				Chain:    model.ChainOUTPUT,
				Protocol: model.ProtocolUDP,
				Action:   model.ActionACCEPT,
				DPort:    "53",
				SIP:      "192.168.1.0/24",
				DIP:      "8.8.8.8",
				Black:    true,
			},
			want: "agent -m=insert -c=OUTPUT -p=udp -a=ACCEPT --dport=53 --sip=192.168.1.0/24 --dip=8.8.8.8 --black",
		},
		{
			name: "rule with TCP flags",
			rule: &model.FirewallRule{
				Chain:    model.ChainINPUT,
				Protocol: model.ProtocolTCP,
				Action:   model.ActionDROP,
				DPort:    "80",
				Options:  &model.ProtocolOptions{TCPFlags: "syn/syn"},
			},
			want: "agent -m=insert -c=INPUT -p=tcp?flags=syn/syn -a=DROP --dport=80",
		},
		{
			name: "rule with ICMP type",
			rule: &model.FirewallRule{
				Chain:    model.ChainINPUT,
				Protocol: model.ProtocolICMP,
				Action:   model.ActionDROP,
				Options:  &model.ProtocolOptions{ICMPType: "echo-request"},
			},
			want: "agent -m=insert -c=INPUT -p=icmp?type=echo-request -a=DROP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RuleToLine(tt.rule)
			if result != tt.want {
				t.Errorf("RuleToLine() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestParseTextToRules 전체 텍스트 파싱 테스트
func TestParseTextToRules(t *testing.T) {
	input := `# Firewall rules
agent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP
agent -m=insert -c=INPUT -p=tcp --dport=9020 -a=DROP

# Another comment
agent -m=insert -c=OUTPUT -p=udp -a=ACCEPT`

	rules, comments, errors := ParseTextToRules(input)

	if len(errors) > 0 {
		t.Errorf("ParseTextToRules() errors: %v", errors)
	}

	if len(rules) != 3 {
		t.Errorf("ParseTextToRules() rules count = %d, want 3", len(rules))
	}

	if len(comments) != 2 {
		t.Errorf("ParseTextToRules() comments count = %d, want 2", len(comments))
	}

	// 첫 번째 규칙 확인
	if rules[0].DPort != "9010" {
		t.Errorf("ParseTextToRules() first rule DPort = %s, want 9010", rules[0].DPort)
	}
}

// TestRulesToText 규칙 목록 → 텍스트 변환 테스트
func TestRulesToText(t *testing.T) {
	rules := []*model.FirewallRule{
		{
			Chain:    model.ChainINPUT,
			Protocol: model.ProtocolTCP,
			Action:   model.ActionDROP,
			DPort:    "9010",
		},
		{
			Chain:    model.ChainINPUT,
			Protocol: model.ProtocolTCP,
			Action:   model.ActionDROP,
			DPort:    "9020",
		},
	}
	comments := []string{"# Firewall rules"}

	result := RulesToText(rules, comments)

	expected := `# Firewall rules
agent -m=insert -c=INPUT -p=tcp -a=DROP --dport=9010
agent -m=insert -c=INPUT -p=tcp -a=DROP --dport=9020`

	if result != expected {
		t.Errorf("RulesToText() = \n%v\n\nwant:\n%v", result, expected)
	}
}

// TestRoundTrip 왕복 변환 테스트 (파싱 → 포맷 → 파싱)
func TestRoundTrip(t *testing.T) {
	original := "agent -m=insert -c=INPUT -p=tcp?flags=syn,rst,ack,fin/syn --dport=80 -a=DROP --sip=192.168.1.0/24"

	rule, err := ParseLine(original)
	if err != nil {
		t.Fatalf("ParseLine() error: %v", err)
	}

	converted := RuleToLine(rule)
	rule2, err := ParseLine(converted)
	if err != nil {
		t.Fatalf("ParseLine() second pass error: %v", err)
	}

	// 주요 필드 비교
	if rule.Chain != rule2.Chain {
		t.Errorf("RoundTrip Chain mismatch: %v != %v", rule.Chain, rule2.Chain)
	}
	if rule.Protocol != rule2.Protocol {
		t.Errorf("RoundTrip Protocol mismatch: %v != %v", rule.Protocol, rule2.Protocol)
	}
	if rule.Action != rule2.Action {
		t.Errorf("RoundTrip Action mismatch: %v != %v", rule.Action, rule2.Action)
	}
	if rule.DPort != rule2.DPort {
		t.Errorf("RoundTrip DPort mismatch: %v != %v", rule.DPort, rule2.DPort)
	}
	if rule.SIP != rule2.SIP {
		t.Errorf("RoundTrip SIP mismatch: %v != %v", rule.SIP, rule2.SIP)
	}
	if rule.Options.TCPFlags != rule2.Options.TCPFlags {
		t.Errorf("RoundTrip TCPFlags mismatch: %v != %v", rule.Options.TCPFlags, rule2.Options.TCPFlags)
	}
}
