package ui

import (
	"fms/internal/model"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// RuleBuilder 규칙 빌더 패널
type RuleBuilder struct {
	ruleTable      *component.RuleTable
	generalForm    *component.RuleForm       // 일반 규칙 폼
	blackWhiteForm *component.BlackWhiteForm // Black/White 폼
	formTabs       *container.AppTabs        // 폼 전환 탭
	onChange       func()
	comments       []string // 주석 라인 보존

	content *fyne.Container
}

// NewRuleBuilder 새 규칙 빌더 생성
func NewRuleBuilder(onChange func()) *RuleBuilder {
	builder := &RuleBuilder{
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

	// 일반 규칙 추가 폼
	b.generalForm = component.NewRuleForm(func(rule *model.FirewallRule) {
		b.ruleTable.AddRule(rule)
		if b.onChange != nil {
			b.onChange()
		}
	})

	// Black/White 규칙 추가 폼
	b.blackWhiteForm = component.NewBlackWhiteForm(func(rule *model.FirewallRule) {
		b.ruleTable.AddRule(rule)
		if b.onChange != nil {
			b.onChange()
		}
	})

	// 폼 전환 탭
	b.formTabs = container.NewAppTabs(
		container.NewTabItem("일반 규칙", b.generalForm.Content()),
		container.NewTabItem("Black/White", b.blackWhiteForm.Content()),
	)

	// 전체 레이아웃: 테이블 위, 폼 탭 아래
	b.content = container.NewBorder(
		nil,
		b.formTabs,
		nil,
		nil,
		b.ruleTable.Content(),
	)
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
