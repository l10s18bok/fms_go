package component

import (
	"fms/internal/model"
	"fms/internal/parser"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 컬럼 인덱스 상수
const (
	colDelete = iota
	colCommand // INSERT/DELETE
	colChain
	colProto
	colOptions
	colAction
	colPort
	colSIP
	colDIP
	colInIF  // 입력 인터페이스 (-i)
	colOutIF // 출력 인터페이스 (-o)
	colBlack
	colWhite
	colCount // 총 컬럼 수 = 13
)

// 고정 너비 컬럼 (픽셀)
const (
	fixedWidthDelete = 36 // 삭제 버튼
	fixedWidthBlack  = 55 // Black 체크박스
	fixedWidthWhite  = 55 // White 체크박스
)

// RuleTable EditableTable 기반 규칙 테이블
type RuleTable struct {
	widget.BaseWidget
	rules    []*model.FirewallRule
	table    *EditableTable
	onChange func()
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
	config := EditableTableConfig{
		Columns: []EditableTableColumn{
			{Header: "", Width: fixedWidthDelete},
			{Header: "Cmd", WidthRatio: 0.07},
			{Header: "Chain", WidthRatio: 0.09},
			{Header: "Proto", WidthRatio: 0.07},
			{Header: "Options", WidthRatio: 0.11},
			{Header: "Action", WidthRatio: 0.08},
			{Header: "Port", WidthRatio: 0.07},
			{Header: "SIP", WidthRatio: 0.15},
			{Header: "DIP", WidthRatio: 0.15},
			{Header: "InIF", WidthRatio: 0.06},
			{Header: "OutIF", WidthRatio: 0.06},
			{Header: "Black", Width: fixedWidthBlack},
			{Header: "White", Width: fixedWidthWhite},
		},
		GetRowCount: func() int {
			return len(t.rules)
		},
		GetCellConfig: func(row, col int) EditableCellConfig {
			return t.getCellConfig(row, col)
		},
	}

	t.table = NewEditableTable(config)
}

// getCellConfig 셀 설정 반환
func (t *RuleTable) getCellConfig(row, col int) EditableCellConfig {
	if row < 0 || row >= len(t.rules) {
		return EditableCellConfig{Type: CellTypeLabel, Text: ""}
	}

	rule := t.rules[row]

	switch col {
	case colDelete:
		return EditableCellConfig{
			Type: CellTypeButton,
			Icon: theme.DeleteIcon(),
			OnTapped: func() {
				t.RemoveRule(row)
			},
		}

	case colCommand:
		return EditableCellConfig{
			Type:     CellTypeSelect,
			Options:  []string{"INSERT", "DELETE"},
			Selected: model.CommandToString(rule.Command),
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].Command = model.StringToCommand(value.(string))
					t.triggerChange()
				}
			},
		}

	case colChain:
		return EditableCellConfig{
			Type:     CellTypeSelect,
			Options:  model.GetChainOptions(),
			Selected: model.ChainToString(rule.Chain),
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].Chain = model.StringToChain(value.(string))
					t.triggerChange()
				}
			},
		}

	case colProto:
		return EditableCellConfig{
			Type:     CellTypeSelect,
			Options:  model.GetProtocolOptions(),
			Selected: model.ProtocolToString(rule.Protocol),
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].Protocol = model.StringToProtocol(value.(string))
					t.rules[row].Options = nil // 프로토콜 변경 시 옵션 초기화
					t.table.Refresh()
					t.triggerChange()
				}
			},
		}

	case colOptions:
		optStr := parser.FormatOptionsOnly(rule.Options)
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: optStr,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].Options = parser.ParseOptionsOnly(value.(string))
					t.triggerChange()
				}
			},
		}

	case colAction:
		return EditableCellConfig{
			Type:     CellTypeSelect,
			Options:  model.GetActionOptions(),
			Selected: model.ActionToString(rule.Action),
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].Action = model.StringToAction(value.(string))
					t.triggerChange()
				}
			},
		}

	case colPort:
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.DPort,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].DPort = value.(string)
					t.triggerChange()
				}
			},
		}

	case colSIP:
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.SIP,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].SIP = value.(string)
					t.triggerChange()
				}
			},
		}

	case colDIP:
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.DIP,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].DIP = value.(string)
					t.triggerChange()
				}
			},
		}

	case colInIF:
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.InInterface,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].InInterface = value.(string)
					t.triggerChange()
				}
			},
		}

	case colOutIF:
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.OutInterface,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].OutInterface = value.(string)
					t.triggerChange()
				}
			},
		}

	case colBlack:
		return EditableCellConfig{
			Type:    CellTypeCheck,
			Checked: rule.Black,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].Black = value.(bool)
					t.triggerChange()
				}
			},
		}

	case colWhite:
		return EditableCellConfig{
			Type:    CellTypeCheck,
			Checked: rule.White,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].White = value.(bool)
					t.triggerChange()
				}
			},
		}
	}

	return EditableCellConfig{Type: CellTypeLabel, Text: ""}
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
