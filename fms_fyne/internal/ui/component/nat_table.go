package component

import (
	"fms/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NAT 테이블 컬럼 인덱스 상수
const (
	natColDelete = iota
	natColCommand   // INSERT/DELETE
	natColType
	natColProto
	natColSrcIP     // MatchIP (-s)
	natColDstIP     // TranslateIP (--dest)
	natColDPort     // MatchPort (--dport)
	natColTransPort // TranslatePort
	natColInIF      // InInterface (-i)
	natColOutIF     // OutInterface (-o)
	natColCount     // 총 컬럼 수 = 10
)

// NAT 테이블 고정 너비 컬럼 (픽셀)
const (
	natFixedWidthDelete  = 36  // 삭제 버튼
	natFixedWidthCommand = 100 // Command (INSERT/DELETE)
	natFixedWidthType    = 120 // NAT 타입 (SNAT/DNAT/MASQUERADE)
	natFixedWidthProto   = 100 // 프로토콜 (TCP/UDP/ANY)
)

// NATTable EditableTable 기반 NAT 규칙 테이블
type NATTable struct {
	widget.BaseWidget
	rules    []*model.NATRule
	table    *EditableTable
	onChange func()
}

// NewNATTable 새 NAT 규칙 테이블 생성
func NewNATTable(onChange func()) *NATTable {
	t := &NATTable{
		rules:    []*model.NATRule{},
		onChange: onChange,
	}
	t.ExtendBaseWidget(t)
	t.createTable()
	return t
}

// createTable 테이블 생성
func (t *NATTable) createTable() {
	config := EditableTableConfig{
		Columns: []EditableTableColumn{
			{Header: "", Width: natFixedWidthDelete},
			{Header: "Cmd", Width: natFixedWidthCommand},
			{Header: "Type", Width: natFixedWidthType},
			{Header: "Proto", Width: natFixedWidthProto},
			{Header: "MatchIP", WidthRatio: 0.18},
			{Header: "TransIP", WidthRatio: 0.18},
			{Header: "MatchPort", WidthRatio: 0.10},
			{Header: "TransPort", WidthRatio: 0.10},
			{Header: "InIF", WidthRatio: 0.08},
			{Header: "OutIF", WidthRatio: 0.08},
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
func (t *NATTable) getCellConfig(row, col int) EditableCellConfig {
	if row < 0 || row >= len(t.rules) {
		return EditableCellConfig{Type: CellTypeLabel, Text: ""}
	}

	rule := t.rules[row]

	switch col {
	case natColDelete:
		return EditableCellConfig{
			Type: CellTypeButton,
			Icon: theme.DeleteIcon(),
			OnTapped: func() {
				t.RemoveRule(row)
			},
		}

	case natColCommand:
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

	case natColType:
		return EditableCellConfig{
			Type:     CellTypeSelect,
			Options:  []string{"SNAT", "DNAT", "MASQUERADE"},
			Selected: model.NATTypeToString(rule.NATType),
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].NATType = model.StringToNATType(value.(string))
					t.triggerChange()
				}
			},
		}

	case natColProto:
		return EditableCellConfig{
			Type:     CellTypeSelect,
			Options:  []string{"tcp", "udp", "any"},
			Selected: model.ProtocolToString(rule.Protocol),
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].Protocol = model.StringToProtocol(value.(string))
					t.triggerChange()
				}
			},
		}

	case natColSrcIP:
		// 소스 IP (-s)
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.MatchIP,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].MatchIP = value.(string)
					t.triggerChange()
				}
			},
		}

	case natColDstIP:
		// 목적지 IP (--dest)
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.TranslateIP,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].TranslateIP = value.(string)
					t.triggerChange()
				}
			},
		}

	case natColDPort:
		// 매칭 포트 (--dport)
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.MatchPort,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].MatchPort = value.(string)
					t.triggerChange()
				}
			},
		}

	case natColTransPort:
		// 변환 포트
		return EditableCellConfig{
			Type: CellTypeEntry,
			Text: rule.TranslatePort,
			OnChanged: func(value any) {
				if row < len(t.rules) {
					t.rules[row].TranslatePort = value.(string)
					t.triggerChange()
				}
			},
		}

	case natColInIF:
		// 입력 인터페이스 (-i)
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

	case natColOutIF:
		// 출력 인터페이스 (-o)
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
	}

	return EditableCellConfig{Type: CellTypeLabel, Text: ""}
}

// CreateRenderer 렌더러 생성
func (t *NATTable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.table)
}

// triggerChange 변경 콜백 호출
func (t *NATTable) triggerChange() {
	if t.onChange != nil {
		t.onChange()
	}
}

// AddRule 규칙 추가
func (t *NATTable) AddRule(rule *model.NATRule) {
	if rule == nil {
		rule = model.NewNATRule()
	}
	t.rules = append(t.rules, rule)
	t.table.Refresh()
	t.triggerChange()
}

// RemoveRule 규칙 삭제
func (t *NATTable) RemoveRule(index int) {
	if index < 0 || index >= len(t.rules) {
		return
	}
	t.rules = append(t.rules[:index], t.rules[index+1:]...)
	t.table.Refresh()
	t.triggerChange()
}

// GetRules 모든 규칙 반환
func (t *NATTable) GetRules() []*model.NATRule {
	return t.rules
}

// SetRules 규칙 목록 설정
func (t *NATTable) SetRules(rules []*model.NATRule) {
	t.rules = rules
	t.table.Refresh()
}

// Clear 목록 초기화
func (t *NATTable) Clear() {
	t.rules = []*model.NATRule{}
	t.table.Refresh()
}

// Content 테이블 위젯 반환
func (t *NATTable) Content() fyne.CanvasObject {
	return t
}

// Refresh 테이블 새로고침
func (t *NATTable) Refresh() {
	t.table.Refresh()
}
