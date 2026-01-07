package component

import (
	"fms/internal/model"
	"fms/internal/parser"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 컬럼 인덱스 상수
const (
	colDelete = iota
	colChain
	colProto
	colOptions
	colAction
	colPort
	colSIP
	colDIP
	colBlack
	colWhite
	colCount // 총 컬럼 수 = 10
)

// 컬럼별 비율 (합계 = 1.0)
var columnRatios = []float32{
	0.04, // 삭제 버튼
	0.10, // Chain
	0.08, // Proto
	0.16, // 옵션
	0.10, // Action
	0.08, // Port
	0.16, // SIP
	0.16, // DIP
	0.06, // Black
	0.06, // White
}

// 헤더 텍스트
var headerTexts = []string{
	"", "Chain", "Proto", "옵션", "Action", "Port", "SIP", "DIP", "B", "W",
}

// RuleTable widget.Table 기반 규칙 테이블
type RuleTable struct {
	widget.BaseWidget
	rules    []*model.FirewallRule
	table    *widget.Table
	onChange func()

	lastWidth float32 // 마지막 너비 (중복 업데이트 방지)
}

// NewRuleTable 새 규칙 테이블 생성
func NewRuleTable(onChange func()) *RuleTable {
	t := &RuleTable{
		rules:    []*model.FirewallRule{},
		onChange: onChange,
	}
	t.ExtendBaseWidget(t)
	t.createTable()
	return t
}

// createTable 테이블 생성
func (t *RuleTable) createTable() {
	t.table = widget.NewTableWithHeaders(
		// Length: 행/열 수 반환
		func() (rows, cols int) {
			return len(t.rules), colCount
		},
		// CreateCell: 셀 위젯 생성
		func() fyne.CanvasObject {
			// 불투명 배경 추가 (Select hover 시 옆 컬럼 텍스트가 비치는 문제 해결)
			bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))

			// 모든 위젯 타입을 Stack에 포함 (컬럼별로 표시/숨김)
			// 인덱스: 0=bg, 1=Button, 2=Select, 3=Entry, 4=Label, 5=Check, 6=Hyperlink
			return container.NewStack(
				bg,
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				widget.NewSelect([]string{}, nil),
				widget.NewEntry(),
				widget.NewLabel(""),
				widget.NewCheck("", nil),
				widget.NewHyperlink("", nil),
			)
		},
		// UpdateCell: 셀 데이터 업데이트
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			t.updateCell(id, obj)
		},
	)

	// 헤더 설정
	t.table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel("")
	}
	t.table.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		label := obj.(*widget.Label)
		if id.Col >= 0 && id.Col < len(headerTexts) {
			label.SetText(headerTexts[id.Col])
		}
	}

	// 초기 컬럼 너비 설정 (기본값)
	t.updateColumnWidths(900)
}

// updateCell 셀 업데이트
func (t *RuleTable) updateCell(id widget.TableCellID, obj fyne.CanvasObject) {
	stack := obj.(*fyne.Container)
	if id.Row < 0 || id.Row >= len(t.rules) {
		return
	}

	rule := t.rules[id.Row]
	row := id.Row

	// 모든 위젯 숨기기
	for _, child := range stack.Objects {
		child.Hide()
	}

	// 배경은 항상 표시 (인덱스 0)
	stack.Objects[0].Show()

	// 인덱스: 0=bg, 1=Button, 2=Select, 3=Entry, 4=Label, 5=Check
	switch id.Col {
	case colDelete:
		btn := stack.Objects[1].(*widget.Button)
		btn.OnTapped = func() {
			t.RemoveRule(row)
		}
		btn.Show()

	case colChain:
		sel := stack.Objects[2].(*widget.Select)
		sel.Options = model.GetChainOptions()
		sel.Selected = model.ChainToString(rule.Chain)
		sel.OnChanged = func(s string) {
			if row < len(t.rules) {
				t.rules[row].Chain = model.StringToChain(s)
				t.triggerChange()
			}
		}
		sel.Show()

	case colProto:
		sel := stack.Objects[2].(*widget.Select)
		sel.Options = model.GetProtocolOptions()
		sel.Selected = model.ProtocolToString(rule.Protocol)
		sel.OnChanged = func(s string) {
			if row < len(t.rules) {
				t.rules[row].Protocol = model.StringToProtocol(s)
				t.rules[row].Options = nil // 프로토콜 변경 시 옵션 초기화
				t.table.Refresh()
				t.triggerChange()
			}
		}
		sel.Show()

	case colOptions:
		link := stack.Objects[6].(*widget.Hyperlink)
		optStr := parser.FormatOptionsOnly(rule.Options)
		if optStr == "" {
			link.SetText("-")
			link.OnTapped = nil
		} else {
			link.SetText(optStr)
			link.OnTapped = func() {
				// 팝업으로 전체 옵션 표시 (오른쪽에 표시)
				popupLabel := widget.NewLabel(optStr)
				popup := widget.NewPopUp(
					container.NewPadded(popupLabel),
					fyne.CurrentApp().Driver().CanvasForObject(link),
				)
				popup.ShowAtRelativePosition(fyne.NewPos(link.Size().Width, 0), link)
			}
		}
		link.Show()

	case colAction:
		sel := stack.Objects[2].(*widget.Select)
		sel.Options = model.GetActionOptions()
		sel.Selected = model.ActionToString(rule.Action)
		sel.OnChanged = func(s string) {
			if row < len(t.rules) {
				t.rules[row].Action = model.StringToAction(s)
				t.triggerChange()
			}
		}
		sel.Show()

	case colPort:
		entry := stack.Objects[3].(*widget.Entry)
		entry.SetText(rule.DPort)
		entry.OnChanged = func(s string) {
			if row < len(t.rules) {
				t.rules[row].DPort = s
				t.triggerChange()
			}
		}
		entry.Show()

	case colSIP:
		entry := stack.Objects[3].(*widget.Entry)
		entry.SetText(rule.SIP)
		entry.OnChanged = func(s string) {
			if row < len(t.rules) {
				t.rules[row].SIP = s
				t.triggerChange()
			}
		}
		entry.Show()

	case colDIP:
		entry := stack.Objects[3].(*widget.Entry)
		entry.SetText(rule.DIP)
		entry.OnChanged = func(s string) {
			if row < len(t.rules) {
				t.rules[row].DIP = s
				t.triggerChange()
			}
		}
		entry.Show()

	case colBlack:
		check := stack.Objects[5].(*widget.Check)
		check.Checked = rule.Black
		check.OnChanged = func(b bool) {
			if row < len(t.rules) {
				t.rules[row].Black = b
				t.triggerChange()
			}
		}
		check.Show()

	case colWhite:
		check := stack.Objects[5].(*widget.Check)
		check.Checked = rule.White
		check.OnChanged = func(b bool) {
			if row < len(t.rules) {
				t.rules[row].White = b
				t.triggerChange()
			}
		}
		check.Show()
	}
}

// updateColumnWidths 컬럼 너비 업데이트 (비율 기반)
func (t *RuleTable) updateColumnWidths(totalWidth float32) {
	if totalWidth <= 0 {
		return
	}
	for i, ratio := range columnRatios {
		t.table.SetColumnWidth(i, totalWidth*ratio)
	}
}

// Resize 크기 변경 시 컬럼 너비 재계산
func (t *RuleTable) Resize(size fyne.Size) {
	if size.Width != t.lastWidth && size.Width > 0 {
		t.lastWidth = size.Width
		t.updateColumnWidths(size.Width)
	}
	t.BaseWidget.Resize(size)
}

// CreateRenderer 렌더러 생성
func (t *RuleTable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.table)
}

// triggerChange 변경 콜백 호출
func (t *RuleTable) triggerChange() {
	if t.onChange != nil {
		t.onChange()
	}
}

// AddRule 규칙 추가
func (t *RuleTable) AddRule(rule *model.FirewallRule) {
	if rule == nil {
		rule = model.NewFirewallRule()
	}
	t.rules = append(t.rules, rule)
	t.table.Refresh()
}

// RemoveRule 규칙 삭제
func (t *RuleTable) RemoveRule(index int) {
	if index < 0 || index >= len(t.rules) {
		return
	}
	t.rules = append(t.rules[:index], t.rules[index+1:]...)
	t.table.Refresh()
	t.triggerChange()
}

// GetRules 모든 규칙 반환
func (t *RuleTable) GetRules() []*model.FirewallRule {
	return t.rules
}

// SetRules 규칙 목록 설정
func (t *RuleTable) SetRules(rules []*model.FirewallRule) {
	t.rules = rules
	t.table.Refresh()
}

// Clear 목록 초기화
func (t *RuleTable) Clear() {
	t.rules = []*model.FirewallRule{}
	t.table.Refresh()
}

// Content 테이블 위젯 반환
func (t *RuleTable) Content() fyne.CanvasObject {
	return t
}

// Refresh 테이블 새로고침
func (t *RuleTable) Refresh() {
	t.table.Refresh()
}
