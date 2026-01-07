package ui

import (
	"fms/internal/model"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NATBuilder NAT 규칙 빌더 패널
type NATBuilder struct {
	natTable *component.NATTable
	dnatForm *component.DNATForm
	snatForm *component.SNATForm
	formTabs *container.AppTabs
	onChange func()
	comments []string // 주석 라인 보존

	content *fyne.Container
}

// NewNATBuilder 새 NAT 빌더 생성
func NewNATBuilder(onChange func()) *NATBuilder {
	builder := &NATBuilder{
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

	// DNAT (포트 포워딩) 폼
	b.dnatForm = component.NewDNATForm(func(rule *model.NATRule) {
		b.natTable.AddRule(rule)
		if b.onChange != nil {
			b.onChange()
		}
	})

	// SNAT/MASQUERADE 폼
	b.snatForm = component.NewSNATForm(func(rule *model.NATRule) {
		b.natTable.AddRule(rule)
		if b.onChange != nil {
			b.onChange()
		}
	})

	// 폼 전환 탭
	b.formTabs = container.NewAppTabs(
		container.NewTabItem("DNAT (포트 포워딩)", b.dnatForm.Content()),
		container.NewTabItem("SNAT/MASQUERADE", b.snatForm.Content()),
	)

	// 전체 레이아웃: 테이블 위, 폼 탭 아래 (Separator로 테이블과 폼 구분)
	formWithSeparator := container.NewVBox(
		widget.NewSeparator(),
		b.formTabs,
	)
	b.content = container.NewBorder(
		nil,
		formWithSeparator,
		nil,
		nil,
		b.natTable.Content(),
	)
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

// ResetTabs 폼 탭 위치 초기화 (첫 번째 탭으로)
func (b *NATBuilder) ResetTabs() {
	if len(b.formTabs.Items) > 0 {
		b.formTabs.SelectIndex(0)
	}
}
