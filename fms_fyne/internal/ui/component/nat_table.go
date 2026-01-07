package component

import (
	"fmt"

	"fms/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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
	// natColDesc  // 설명 컬럼 - 현재 미사용
	natColCount // 총 컬럼 수 = 6
)

// NAT 테이블 고정 너비 컬럼 (픽셀)
const (
	natFixedWidthDelete = 36  // 삭제 버튼
	natFixedWidthType   = 100 // NAT 타입
	natFixedWidthProto  = 70  // 프로토콜
	natScrollbarWidth   = 32  // 스크롤바 및 테이블 여백
)

// NAT 테이블 가변 컬럼별 비율 (합계 = 1.0)
var natColumnRatios = []float32{
	0.30, // Match
	0.35, // Translate
	0.35, // Interface
	// 0.25, // Description - 현재 미사용
}

// NAT 테이블 헤더 텍스트
var natHeaderTexts = []string{
	"", "Type", "Proto", "Match", "Translate", "Interface",
	// "Desc", // 설명 컬럼 - 현재 미사용
}

// NATTable widget.Table 기반 NAT 규칙 테이블
type NATTable struct {
	widget.BaseWidget
	rules    []*model.NATRule
	table    *widget.Table
	onChange func()

	lastWidth float32 // 마지막 너비 (중복 업데이트 방지)
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
	t.table = widget.NewTable(
		// Length: 행/열 수 반환
		func() (rows, cols int) {
			return len(t.rules), natColCount
		},
		// CreateCell: 셀 위젯 생성
		func() fyne.CanvasObject {
			// 불투명 배경 추가
			bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))

			// 모든 위젯 타입을 Stack에 포함
			// 인덱스: 0=bg, 1=Button, 2=Label
			return container.NewStack(
				bg,
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				widget.NewLabel(""),
			)
		},
		// UpdateCell: 셀 데이터 업데이트
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			t.updateCell(id, obj)
		},
	)

	// 헤더 설정 (컬럼 헤더만, 행 번호 헤더 없음)
	t.table.ShowHeaderRow = true
	t.table.ShowHeaderColumn = false
	t.table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel("")
	}
	t.table.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		label := obj.(*widget.Label)
		if id.Col >= 0 && id.Col < len(natHeaderTexts) {
			label.SetText(natHeaderTexts[id.Col])
		}
	}

	// 초기 컬럼 너비 설정 (기본값)
	t.updateColumnWidths(900)
}

// updateCell 셀 업데이트
func (t *NATTable) updateCell(id widget.TableCellID, obj fyne.CanvasObject) {
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

	// 인덱스: 0=bg, 1=Button, 2=Label
	switch id.Col {
	case natColDelete:
		btn := stack.Objects[1].(*widget.Button)
		btn.OnTapped = func() {
			t.RemoveRule(row)
		}
		btn.Show()

	case natColType:
		label := stack.Objects[2].(*widget.Label)
		label.SetText(model.NATTypeToString(rule.NATType))
		label.Show()

	case natColProto:
		label := stack.Objects[2].(*widget.Label)
		label.SetText(model.ProtocolToString(rule.Protocol))
		label.Show()

	case natColMatch:
		label := stack.Objects[2].(*widget.Label)
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
		label.SetText(matchStr)
		label.Show()

	case natColTranslate:
		label := stack.Objects[2].(*widget.Label)
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
		label.SetText(transStr)
		label.Show()

	case natColInterface:
		label := stack.Objects[2].(*widget.Label)
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
		label.SetText(ifStr)
		label.Show()

	// case natColDesc: // 설명 컬럼 - 현재 미사용
	// 	label := stack.Objects[2].(*widget.Label)
	// 	desc := rule.Description
	// 	if desc == "" {
	// 		desc = "-"
	// 	}
	// 	label.SetText(desc)
	// 	label.Show()
	}
}

// updateColumnWidths 컬럼 너비 업데이트 (고정 + 비율 기반)
func (t *NATTable) updateColumnWidths(totalWidth float32) {
	if totalWidth <= 0 {
		return
	}

	// 고정 너비 컬럼 설정
	t.table.SetColumnWidth(natColDelete, natFixedWidthDelete)
	t.table.SetColumnWidth(natColType, natFixedWidthType)
	t.table.SetColumnWidth(natColProto, natFixedWidthProto)

	// 가변 너비 계산 (전체 - 고정 컬럼들 - 스크롤바 너비)
	flexibleWidth := totalWidth - natFixedWidthDelete - natFixedWidthType - natFixedWidthProto - natScrollbarWidth

	// 가변 컬럼에 비율 적용
	flexibleCols := []int{natColMatch, natColTranslate, natColInterface}
	for i, col := range flexibleCols {
		if i < len(natColumnRatios) {
			t.table.SetColumnWidth(col, flexibleWidth*natColumnRatios[i])
		}
	}
}

// Resize 크기 변경 시 컬럼 너비 재계산
func (t *NATTable) Resize(size fyne.Size) {
	if size.Width != t.lastWidth && size.Width > 0 {
		t.lastWidth = size.Width
		t.updateColumnWidths(size.Width)
	}
	t.BaseWidget.Resize(size)
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
