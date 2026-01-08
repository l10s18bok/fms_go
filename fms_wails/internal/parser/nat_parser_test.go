package parser

import (
	"strings"
	"testing"

	"fms_wails/internal/model"
)

// TestParseNATLine_DNAT DNAT 규칙 파싱 테스트
func TestParseNATLine_DNAT(t *testing.T) {
	line := "agent -m=insert -t=nat --nat-type=dnat -p=tcp --match-port=6080 --to-dest=192.168.30.180:8080"

	rule, err := ParseNATLine(line)
	if err != nil {
		t.Fatalf("ParseNATLine() error = %v", err)
	}

	if rule.NATType != model.NATTypeDNAT {
		t.Errorf("NATType = %d, want %d", rule.NATType, model.NATTypeDNAT)
	}
	if rule.Protocol != model.ProtocolTCP {
		t.Errorf("Protocol = %d, want %d", rule.Protocol, model.ProtocolTCP)
	}
	if rule.MatchPort != "6080" {
		t.Errorf("MatchPort = %s, want 6080", rule.MatchPort)
	}
	if rule.TranslateIP != "192.168.30.180" {
		t.Errorf("TranslateIP = %s, want 192.168.30.180", rule.TranslateIP)
	}
	if rule.TranslatePort != "8080" {
		t.Errorf("TranslatePort = %s, want 8080", rule.TranslatePort)
	}
}

// TestParseNATLine_SNAT SNAT 규칙 파싱 테스트
func TestParseNATLine_SNAT(t *testing.T) {
	line := "agent -m=insert -t=nat --nat-type=snat -p=tcp -s=192.168.1.0/24 --to-source=203.0.113.1 -o=eth0"

	rule, err := ParseNATLine(line)
	if err != nil {
		t.Fatalf("ParseNATLine() error = %v", err)
	}

	if rule.NATType != model.NATTypeSNAT {
		t.Errorf("NATType = %d, want %d", rule.NATType, model.NATTypeSNAT)
	}
	if rule.MatchIP != "192.168.1.0/24" {
		t.Errorf("MatchIP = %s, want 192.168.1.0/24", rule.MatchIP)
	}
	if rule.TranslateIP != "203.0.113.1" {
		t.Errorf("TranslateIP = %s, want 203.0.113.1", rule.TranslateIP)
	}
	if rule.OutInterface != "eth0" {
		t.Errorf("OutInterface = %s, want eth0", rule.OutInterface)
	}
}

// TestParseNATLine_MASQUERADE MASQUERADE 규칙 파싱 테스트
func TestParseNATLine_MASQUERADE(t *testing.T) {
	line := "agent -m=insert -t=nat --nat-type=masquerade -p=any -s=192.168.1.0/24 -o=eth0"

	rule, err := ParseNATLine(line)
	if err != nil {
		t.Fatalf("ParseNATLine() error = %v", err)
	}

	if rule.NATType != model.NATTypeMASQUERADE {
		t.Errorf("NATType = %d, want %d", rule.NATType, model.NATTypeMASQUERADE)
	}
	if rule.MatchIP != "192.168.1.0/24" {
		t.Errorf("MatchIP = %s, want 192.168.1.0/24", rule.MatchIP)
	}
	if rule.OutInterface != "eth0" {
		t.Errorf("OutInterface = %s, want eth0", rule.OutInterface)
	}
}

// TestParseNATLine_EmptyLine 빈 라인 테스트
func TestParseNATLine_EmptyLine(t *testing.T) {
	rule, err := ParseNATLine("")
	if err != nil {
		t.Errorf("ParseNATLine('') should not return error")
	}
	if rule != nil {
		t.Errorf("ParseNATLine('') should return nil")
	}
}

// TestParseNATLine_CommentLine 주석 라인 테스트
func TestParseNATLine_CommentLine(t *testing.T) {
	rule, err := ParseNATLine("# 이것은 주석입니다")
	if err != nil {
		t.Errorf("ParseNATLine(comment) should not return error")
	}
	if rule != nil {
		t.Errorf("ParseNATLine(comment) should return nil")
	}
}

// TestParseNATLine_InvalidFormat 잘못된 형식 테스트
func TestParseNATLine_InvalidFormat(t *testing.T) {
	_, err := ParseNATLine("invalid line")
	if err == nil {
		t.Error("ParseNATLine(invalid) should return error")
	}
}

// TestParseNATLine_NotNAT NAT이 아닌 규칙 테스트
func TestParseNATLine_NotNAT(t *testing.T) {
	_, err := ParseNATLine("agent -m=insert -t=filter -p=tcp")
	if err == nil {
		t.Error("ParseNATLine(filter) should return error")
	}
}

// TestNATRuleToLine_DNAT DNAT 규칙 변환 테스트
func TestNATRuleToLine_DNAT(t *testing.T) {
	rule := &model.NATRule{
		NATType:       model.NATTypeDNAT,
		Protocol:      model.ProtocolTCP,
		MatchPort:     "6080",
		TranslateIP:   "192.168.30.180",
		TranslatePort: "8080",
	}

	line := NATRuleToLine(rule)

	if !strings.Contains(line, "--nat-type=dnat") {
		t.Error("Line should contain --nat-type=dnat")
	}
	if !strings.Contains(line, "--match-port=6080") {
		t.Error("Line should contain --match-port=6080")
	}
	if !strings.Contains(line, "--to-dest=192.168.30.180:8080") {
		t.Error("Line should contain --to-dest=192.168.30.180:8080")
	}
}

// TestNATRuleToLine_SNAT SNAT 규칙 변환 테스트
func TestNATRuleToLine_SNAT(t *testing.T) {
	rule := &model.NATRule{
		NATType:      model.NATTypeSNAT,
		Protocol:     model.ProtocolTCP,
		MatchIP:      "192.168.1.0/24",
		TranslateIP:  "203.0.113.1",
		OutInterface: "eth0",
	}

	line := NATRuleToLine(rule)

	if !strings.Contains(line, "--nat-type=snat") {
		t.Error("Line should contain --nat-type=snat")
	}
	if !strings.Contains(line, "-s=192.168.1.0/24") {
		t.Error("Line should contain -s=192.168.1.0/24")
	}
	if !strings.Contains(line, "--to-source=203.0.113.1") {
		t.Error("Line should contain --to-source=203.0.113.1")
	}
}

// TestNATRuleToLine_Nil nil 입력 테스트
func TestNATRuleToLine_Nil(t *testing.T) {
	line := NATRuleToLine(nil)
	if line != "" {
		t.Error("NATRuleToLine(nil) should return empty string")
	}
}

// TestNATRuleToSmartfw_DNAT smartfw DNAT 변환 테스트
func TestNATRuleToSmartfw_DNAT(t *testing.T) {
	rule := &model.NATRule{
		NATType:       model.NATTypeDNAT,
		Protocol:      model.ProtocolTCP,
		MatchIP:       "ANY",
		MatchPort:     "6080",
		TranslateIP:   "192.168.30.180",
		TranslatePort: "8080",
	}

	result := NATRuleToSmartfw(rule, "123456")

	expected := "req|INSERT|123456|ANY|NAT|ANY|TCP?DNAT|192.168.30.180|6080,8080||"
	if result != expected {
		t.Errorf("NATRuleToSmartfw() = %s, want %s", result, expected)
	}
}

// TestNATRuleToSmartfw_SNAT smartfw SNAT 변환 테스트
func TestNATRuleToSmartfw_SNAT(t *testing.T) {
	rule := &model.NATRule{
		NATType:      model.NATTypeSNAT,
		Protocol:     model.ProtocolTCP,
		MatchIP:      "192.168.1.0/24",
		TranslateIP:  "203.0.113.1",
		OutInterface: "eth0",
	}

	result := NATRuleToSmartfw(rule, "123456")

	if !strings.Contains(result, "TCP?SNAT") {
		t.Error("Result should contain TCP?SNAT")
	}
	if !strings.Contains(result, "192.168.1.0/24") {
		t.Error("Result should contain 192.168.1.0/24")
	}
}

// TestNATRuleToSmartfw_MASQUERADE smartfw MASQUERADE 변환 테스트
func TestNATRuleToSmartfw_MASQUERADE(t *testing.T) {
	rule := &model.NATRule{
		NATType:      model.NATTypeMASQUERADE,
		Protocol:     model.ProtocolANY,
		MatchIP:      "192.168.1.0/24",
		OutInterface: "eth0",
	}

	result := NATRuleToSmartfw(rule, "123456")

	if !strings.Contains(result, "ANY?MASQUERADE") {
		t.Error("Result should contain ANY?MASQUERADE")
	}
}

// TestParseTextToNATRules 텍스트 파싱 테스트
func TestParseTextToNATRules(t *testing.T) {
	text := `# NAT 규칙
agent -m=insert -t=nat --nat-type=dnat -p=tcp --match-port=80 --to-dest=192.168.1.100:80
agent -m=insert -t=nat --nat-type=snat -p=tcp -s=192.168.1.0/24 --to-source=203.0.113.1
# MASQUERADE
agent -m=insert -t=nat --nat-type=masquerade -p=any -s=10.0.0.0/8 -o=eth0`

	rules, comments, errors := ParseTextToNATRules(text)

	if len(errors) > 0 {
		t.Errorf("ParseTextToNATRules() errors = %v", errors)
	}

	if len(rules) != 3 {
		t.Errorf("ParseTextToNATRules() rules count = %d, want 3", len(rules))
	}

	if len(comments) != 2 {
		t.Errorf("ParseTextToNATRules() comments count = %d, want 2", len(comments))
	}

	// 첫 번째 규칙 확인
	if rules[0].NATType != model.NATTypeDNAT {
		t.Errorf("First rule NATType = %d, want %d", rules[0].NATType, model.NATTypeDNAT)
	}

	// 두 번째 규칙 확인
	if rules[1].NATType != model.NATTypeSNAT {
		t.Errorf("Second rule NATType = %d, want %d", rules[1].NATType, model.NATTypeSNAT)
	}

	// 세 번째 규칙 확인
	if rules[2].NATType != model.NATTypeMASQUERADE {
		t.Errorf("Third rule NATType = %d, want %d", rules[2].NATType, model.NATTypeMASQUERADE)
	}
}

// TestNATRulesToText 규칙 → 텍스트 변환 테스트
func TestNATRulesToText(t *testing.T) {
	rules := []*model.NATRule{
		{
			NATType:       model.NATTypeDNAT,
			Protocol:      model.ProtocolTCP,
			MatchPort:     "80",
			TranslateIP:   "192.168.1.100",
			TranslatePort: "80",
		},
	}
	comments := []string{"# NAT 규칙"}

	text := NATRulesToText(rules, comments)

	if !strings.Contains(text, "# NAT 규칙") {
		t.Error("Text should contain comment")
	}
	if !strings.Contains(text, "--nat-type=dnat") {
		t.Error("Text should contain DNAT rule")
	}
}

// TestRoundTrip 왕복 변환 테스트
func TestNATRoundTrip(t *testing.T) {
	original := &model.NATRule{
		NATType:       model.NATTypeDNAT,
		Protocol:      model.ProtocolTCP,
		MatchPort:     "443",
		MatchIP:       "10.0.0.0/8",
		TranslateIP:   "192.168.1.100",
		TranslatePort: "8443",
	}

	// 규칙 → 텍스트
	line := NATRuleToLine(original)

	// 텍스트 → 규칙
	parsed, err := ParseNATLine(line)
	if err != nil {
		t.Fatalf("ParseNATLine() error = %v", err)
	}

	// 비교
	if parsed.NATType != original.NATType {
		t.Errorf("NATType mismatch: %d != %d", parsed.NATType, original.NATType)
	}
	if parsed.Protocol != original.Protocol {
		t.Errorf("Protocol mismatch: %d != %d", parsed.Protocol, original.Protocol)
	}
	if parsed.MatchPort != original.MatchPort {
		t.Errorf("MatchPort mismatch: %s != %s", parsed.MatchPort, original.MatchPort)
	}
	if parsed.TranslateIP != original.TranslateIP {
		t.Errorf("TranslateIP mismatch: %s != %s", parsed.TranslateIP, original.TranslateIP)
	}
	if parsed.TranslatePort != original.TranslatePort {
		t.Errorf("TranslatePort mismatch: %s != %s", parsed.TranslatePort, original.TranslatePort)
	}
}

// TestIsNATLine NAT 라인 확인 테스트
func TestIsNATLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"agent -m=insert -t=nat --nat-type=dnat", true},
		{"agent -m=insert -t=filter -p=tcp", false},
		{"# comment", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsNATLine(tt.line)
		if result != tt.expected {
			t.Errorf("IsNATLine(%s) = %v, want %v", tt.line, result, tt.expected)
		}
	}
}
