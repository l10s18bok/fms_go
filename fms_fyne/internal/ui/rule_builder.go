package ui

import (
	"fms/internal/model"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// RuleBuilder 룰 빌더 패널
type RuleBuilder struct {
	window    fyne.Window
	ruleTable *component.RuleTable
	addDialog *component.RuleAddDialog // 규칙 추가 다이얼로그
	onChange  func()
	comments  []string // 주석 라인 보존

	content *fyne.Container
}

// NewRuleBuilder 새 룰 빌더 생성
func NewRuleBuilder(window fyne.Window, onChange func()) *RuleBuilder {
	builder := &RuleBuilder{
		window:   window,
		onChange: onChange,
		comments: []string{},
	}
	builder.createUI()
	return builder
}

// createUI UI 생성
func (b *RuleBuilder) createUI() {
	// 규칙 테이블
	b.ruleTable = component.NewRuleTable(b.onChange)

	// 규칙 추가 다이얼로그
	b.addDialog = component.NewRuleAddDialog(b.window, func(rule *model.FirewallRule) {
		b.ruleTable.AddRule(rule)
		if b.onChange != nil {
			b.onChange()
		}
	})

	// 전체 레이아웃: 테이블만 표시 (추가 버튼은 TemplateTab 헤더로 이동)
	b.content = container.NewMax(b.ruleTable.Content())
}

// Content UI 컨테이너 반환
func (b *RuleBuilder) Content() *fyne.Container {
	return b.content
}

// GetRules 모든 규칙 반환
func (b *RuleBuilder) GetRules() []*model.FirewallRule {
	return b.ruleTable.GetRules()
}

// SetRules 규칙 목록 설정
func (b *RuleBuilder) SetRules(rules []*model.FirewallRule) {
	b.ruleTable.SetRules(rules)
}

// GetComments 주석 반환
func (b *RuleBuilder) GetComments() []string {
	return b.comments
}

// SetComments 주석 설정
func (b *RuleBuilder) SetComments(comments []string) {
	b.comments = comments
}

// Clear 초기화
func (b *RuleBuilder) Clear() {
	b.ruleTable.Clear()
	b.comments = []string{}
}

// Refresh UI 새로고침
func (b *RuleBuilder) Refresh() {
	b.ruleTable.Refresh()
}

// ResetTabs 다이얼로그 탭 위치 초기화 (첫 번째 탭으로)
func (b *RuleBuilder) ResetTabs() {
	b.addDialog.ResetTabs()
}

// ShowAddDialog 규칙 추가 다이얼로그 표시
func (b *RuleBuilder) ShowAddDialog() {
	b.addDialog.Show()
}
