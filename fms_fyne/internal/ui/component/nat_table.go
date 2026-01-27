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
	natColType
	natColProto
	natColSrcIP     // MatchIP (-s)
	natColDstIP     // TranslateIP (--dest)
	natColDPort     // MatchPort (--dport)
	natColTransPort // TranslatePort
	natColInIF      // InInterface (-i)
	natColOutIF     // OutInterface (-o)
	natColCount     // 총 컬럼 수 = 9
)

// NAT 테이블 고정 너비 컬럼 (픽셀)
const (
	natFixedWidthDelete = 36  // 삭제 버튼
	natFixedWidthType   = 100 // NAT 타입
	natFixedWidthProto  = 70  // 프로토콜
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

	case natColType:
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: model.NATTypeToString(rule.NATType),
		}

	case natColProto:
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: model.ProtocolToString(rule.Protocol),
		}

	case natColSrcIP:
		// 소스 IP (-s)
		srcIP := rule.MatchIP
		if srcIP == "" {
			srcIP = "-"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: srcIP,
		}

	case natColDstIP:
		// 목적지 IP (--dest)
		dstIP := rule.TranslateIP
		if dstIP == "" {
			dstIP = "-"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: dstIP,
		}

	case natColDPort:
		// 매칭 포트 (--dport)
		dport := rule.MatchPort
		if dport == "" {
			dport = "-"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: dport,
		}

	case natColTransPort:
		// 변환 포트
		tport := rule.TranslatePort
		if tport == "" {
			tport = "-"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: tport,
		}

	case natColInIF:
		// 입력 인터페이스 (-i)
		inIF := rule.InInterface
		if inIF == "" {
			inIF = "-"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: inIF,
		}

	case natColOutIF:
		// 출력 인터페이스 (-o)
		outIF := rule.OutInterface
		if outIF == "" {
			outIF = "-"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: outIF,
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
