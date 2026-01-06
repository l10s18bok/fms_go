package component

import (
	"fms/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// RuleList 규칙 목록 컴포넌트
type RuleList struct {
	rows     []*RuleRow
	onChange func()

	header    *fyne.Container
	rowsBox   *fyne.Container
	scroll    *container.Scroll
	content   *fyne.Container
}

// NewRuleList 새 규칙 목록 생성
func NewRuleList(onChange func()) *RuleList {
	list := &RuleList{
		rows:     []*RuleRow{},
		onChange: onChange,
	}
	list.createUI()
	return list
}

// createUI UI 생성
func (l *RuleList) createUI() {
	// 헤더 행 생성 (옵션 컬럼 추가)
	l.header = container.NewHBox(
		container.NewGridWrap(fyne.NewSize(36, 30), widget.NewLabel("")),       // 삭제 버튼 자리
		container.NewGridWrap(fyne.NewSize(100, 30), widget.NewLabel("Chain")),
		container.NewGridWrap(fyne.NewSize(80, 30), widget.NewLabel("Proto")),
		container.NewGridWrap(fyne.NewSize(150, 30), widget.NewLabel("옵션")),   // 옵션 컬럼
		container.NewGridWrap(fyne.NewSize(90, 30), widget.NewLabel("Action")),
		container.NewGridWrap(fyne.NewSize(80, 30), widget.NewLabel("Port")),
		container.NewGridWrap(fyne.NewSize(140, 30), widget.NewLabel("SIP")),
		container.NewGridWrap(fyne.NewSize(140, 30), widget.NewLabel("DIP")),
		container.NewGridWrap(fyne.NewSize(30, 30), widget.NewLabel("B")),
		container.NewGridWrap(fyne.NewSize(30, 30), widget.NewLabel("W")),
	)

	// 규칙 행들을 담을 VBox
	l.rowsBox = container.NewVBox()

	// 스크롤 컨테이너
	l.scroll = container.NewVScroll(l.rowsBox)
	l.scroll.SetMinSize(fyne.NewSize(700, 200))

	// 전체 레이아웃
	l.content = container.NewBorder(l.header, nil, nil, nil, l.scroll)
}

// AddRule 규칙 추가
func (l *RuleList) AddRule(rule *model.FirewallRule) {
	index := len(l.rows)
	row := NewRuleRow(rule, func() {
		l.removeRowAt(index)
	}, l.onChange)

	l.rows = append(l.rows, row)
	l.rowsBox.Add(row.Content())
	l.rowsBox.Refresh()

	// 삭제 콜백 인덱스 갱신
	l.updateDeleteCallbacks()
}

// removeRowAt 특정 인덱스의 행 삭제
func (l *RuleList) removeRowAt(index int) {
	if index < 0 || index >= len(l.rows) {
		return
	}

	// 행 제거
	l.rows = append(l.rows[:index], l.rows[index+1:]...)

	// UI 재구성
	l.rebuildRowsBox()

	if l.onChange != nil {
		l.onChange()
	}
}

// updateDeleteCallbacks 삭제 콜백 인덱스 갱신
func (l *RuleList) updateDeleteCallbacks() {
	for i := range l.rows {
		index := i // 클로저를 위한 로컬 변수
		l.rows[i].onDelete = func() {
			l.removeRowAt(index)
		}
	}
}

// rebuildRowsBox 행 UI 재구성
func (l *RuleList) rebuildRowsBox() {
	l.rowsBox.Objects = nil
	for _, row := range l.rows {
		l.rowsBox.Add(row.Content())
	}
	l.rowsBox.Refresh()
	l.updateDeleteCallbacks()
}

// GetRules 모든 규칙 반환
func (l *RuleList) GetRules() []*model.FirewallRule {
	rules := make([]*model.FirewallRule, len(l.rows))
	for i, row := range l.rows {
		rules[i] = row.GetRule()
	}
	return rules
}

// SetRules 규칙 목록 설정
func (l *RuleList) SetRules(rules []*model.FirewallRule) {
	l.rows = []*RuleRow{}
	l.rowsBox.Objects = nil

	for _, rule := range rules {
		row := NewRuleRow(rule, nil, l.onChange)
		l.rows = append(l.rows, row)
		l.rowsBox.Add(row.Content())
	}

	l.updateDeleteCallbacks()
	l.rowsBox.Refresh()
}

// Clear 목록 초기화
func (l *RuleList) Clear() {
	l.rows = []*RuleRow{}
	l.rowsBox.Objects = nil
	l.rowsBox.Refresh()
}

// Content UI 컨테이너 반환
func (l *RuleList) Content() *fyne.Container {
	return l.content
}

// Refresh UI 새로고침
func (l *RuleList) Refresh() {
	l.rowsBox.Refresh()
}
