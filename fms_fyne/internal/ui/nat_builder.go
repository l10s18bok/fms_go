package ui

import (
	"fms/internal/model"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// NATBuilder NAT 규칙 빌더 패널
type NATBuilder struct {
	window    fyne.Window
	natTable  *component.NATTable
	addDialog *component.NATAddDialog // NAT 규칙 추가 다이얼로그
	onChange  func()
	comments  []string // 주석 라인 보존

	content *fyne.Container
}

// NewNATBuilder 새 NAT 빌더 생성
func NewNATBuilder(window fyne.Window, onChange func()) *NATBuilder {
	builder := &NATBuilder{
		window:   window,
		onChange: onChange,
		comments: []string{},
	}
	builder.createUI()
	return builder
}

// createUI UI 생성
func (b *NATBuilder) createUI() {
	// NAT 규칙 테이블
	b.natTable = component.NewNATTable(b.onChange)

	// NAT 규칙 추가 다이얼로그
	b.addDialog = component.NewNATAddDialog(b.window, func(rule *model.NATRule) {
		b.natTable.AddRule(rule)
		if b.onChange != nil {
			b.onChange()
		}
	})

	// 전체 레이아웃: 테이블만 표시 (추가 버튼은 TemplateTab 헤더로 이동)
	b.content = container.NewMax(b.natTable.Content())
}

// Content UI 컨테이너 반환
func (b *NATBuilder) Content() *fyne.Container {
	return b.content
}

// GetRules 모든 NAT 규칙 반환
func (b *NATBuilder) GetRules() []*model.NATRule {
	return b.natTable.GetRules()
}

// SetRules NAT 규칙 목록 설정
func (b *NATBuilder) SetRules(rules []*model.NATRule) {
	b.natTable.SetRules(rules)
}

// GetComments 주석 반환
func (b *NATBuilder) GetComments() []string {
	return b.comments
}

// SetComments 주석 설정
func (b *NATBuilder) SetComments(comments []string) {
	b.comments = comments
}

// Clear 초기화
func (b *NATBuilder) Clear() {
	b.natTable.Clear()
	b.comments = []string{}
}

// Refresh UI 새로고침
func (b *NATBuilder) Refresh() {
	b.natTable.Refresh()
}

// ResetTabs 다이얼로그 탭 위치 초기화 (첫 번째 탭으로)
func (b *NATBuilder) ResetTabs() {
	b.addDialog.ResetTabs()
}

// ShowAddDialog NAT 규칙 추가 다이얼로그 표시
func (b *NATBuilder) ShowAddDialog() {
	b.addDialog.Show()
}
