package component

import (
	"fmt"

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
	natColMatch
	natColTranslate
	natColInterface
	natColCount // 총 컬럼 수 = 6
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
			{Header: "Match", WidthRatio: 0.30},
			{Header: "Translate", WidthRatio: 0.35},
			{Header: "Interface", WidthRatio: 0.35},
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

	case natColMatch:
		// 매칭 조건 표시: IP:Port 형식
		matchStr := ""
		if rule.MatchIP != "" && rule.MatchIP != "ANY" {
			matchStr = rule.MatchIP
		}
		if rule.MatchPort != "" {
			if matchStr != "" {
				matchStr += ":"
			}
			matchStr += rule.MatchPort
		}
		if matchStr == "" {
			matchStr = "ANY"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: matchStr,
		}

	case natColTranslate:
		// 변환 대상 표시: IP:Port 형식
		transStr := ""
		if rule.TranslateIP != "" {
			transStr = rule.TranslateIP
		}
		if rule.TranslatePort != "" {
			if transStr != "" {
				transStr += ":"
			}
			transStr += rule.TranslatePort
		}
		if transStr == "" {
			transStr = "-"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: transStr,
		}

	case natColInterface:
		// 인터페이스 표시: IN/OUT 형식
		ifStr := ""
		if rule.InInterface != "" {
			ifStr = fmt.Sprintf("IN:%s", rule.InInterface)
		}
		if rule.OutInterface != "" {
			if ifStr != "" {
				ifStr += " "
			}
			ifStr += fmt.Sprintf("OUT:%s", rule.OutInterface)
		}
		if ifStr == "" {
			ifStr = "-"
		}
		return EditableCellConfig{
			Type: CellTypeLabel,
			Text: ifStr,
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
