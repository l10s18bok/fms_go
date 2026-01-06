package component

import (
	"fms/internal/model"
	"fms/internal/themes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// RuleRow 규칙 테이블의 한 행을 나타내는 컴포넌트
type RuleRow struct {
	rule     *model.FirewallRule
	onDelete func()
	onChange func()

	// UI 요소
	deleteBtn   fyne.CanvasObject
	chainSel    *widget.Select
	protoSel    *widget.Select
	actionSel   *widget.Select
	dportEntry  *widget.Entry
	sipEntry    *widget.Entry
	dipEntry    *widget.Entry
	blackCheck  *widget.Check
	whiteCheck  *widget.Check
	content     *fyne.Container
}

// NewRuleRow 새 규칙 행 생성
func NewRuleRow(rule *model.FirewallRule, onDelete, onChange func()) *RuleRow {
	if rule == nil {
		rule = model.NewFirewallRule()
	}

	row := &RuleRow{
		rule:     rule,
		onDelete: onDelete,
		onChange: onChange,
	}
	row.createUI()
	row.syncFromRule()
	return row
}

// createUI UI 요소 생성
func (r *RuleRow) createUI() {
	// 삭제 버튼 (아이콘만, 어두운 회색 배경, 좌우 패딩 5)
	r.deleteBtn = NewCustomButton("", theme.DeleteIcon(), nil, themes.Colors["darkgray"], func() {
		if r.onDelete != nil {
			r.onDelete()
		}
	}, 0, 0, 5, 5)

	// Chain 선택
	r.chainSel = widget.NewSelect(model.GetChainOptions(), func(s string) {
		r.rule.Chain = model.StringToChain(s)
		r.triggerChange()
	})

	// Protocol 선택
	r.protoSel = widget.NewSelect(model.GetProtocolOptions(), func(s string) {
		r.rule.Protocol = model.StringToProtocol(s)
		r.triggerChange()
	})

	// Action 선택
	r.actionSel = widget.NewSelect(model.GetActionOptions(), func(s string) {
		r.rule.Action = model.StringToAction(s)
		r.triggerChange()
	})

	// DPort 입력
	r.dportEntry = widget.NewEntry()
	r.dportEntry.SetPlaceHolder("포트")
	r.dportEntry.OnChanged = func(s string) {
		r.rule.DPort = s
		r.triggerChange()
	}

	// SIP 입력
	r.sipEntry = widget.NewEntry()
	r.sipEntry.SetPlaceHolder("Source IP")
	r.sipEntry.OnChanged = func(s string) {
		r.rule.SIP = s
		r.triggerChange()
	}

	// DIP 입력
	r.dipEntry = widget.NewEntry()
	r.dipEntry.SetPlaceHolder("Dest IP")
	r.dipEntry.OnChanged = func(s string) {
		r.rule.DIP = s
		r.triggerChange()
	}

	// 체크박스들
	r.blackCheck = widget.NewCheck("", func(b bool) {
		r.rule.Black = b
		r.triggerChange()
	})
	r.whiteCheck = widget.NewCheck("", func(b bool) {
		r.rule.White = b
		r.triggerChange()
	})

	// 레이아웃 구성
	r.content = container.NewHBox(
		container.NewGridWrap(fyne.NewSize(36, 36), r.deleteBtn),
		container.NewGridWrap(fyne.NewSize(110, 36), r.chainSel),
		container.NewGridWrap(fyne.NewSize(110, 36), r.protoSel),
		container.NewGridWrap(fyne.NewSize(110, 36), r.actionSel),
		container.NewGridWrap(fyne.NewSize(110, 36), r.dportEntry),
		container.NewGridWrap(fyne.NewSize(180, 36), r.sipEntry),
		container.NewGridWrap(fyne.NewSize(180, 36), r.dipEntry),
		container.NewGridWrap(fyne.NewSize(30, 36), r.blackCheck),
		container.NewGridWrap(fyne.NewSize(30, 36), r.whiteCheck),
	)
}

// syncFromRule 규칙 데이터를 UI에 반영
func (r *RuleRow) syncFromRule() {
	r.chainSel.SetSelected(model.ChainToString(r.rule.Chain))
	r.protoSel.SetSelected(model.ProtocolToString(r.rule.Protocol))
	r.actionSel.SetSelected(model.ActionToString(r.rule.Action))
	r.dportEntry.SetText(r.rule.DPort)
	r.sipEntry.SetText(r.rule.SIP)
	r.dipEntry.SetText(r.rule.DIP)
	r.blackCheck.SetChecked(r.rule.Black)
	r.whiteCheck.SetChecked(r.rule.White)
}

// triggerChange 변경 콜백 호출
func (r *RuleRow) triggerChange() {
	if r.onChange != nil {
		r.onChange()
	}
}

// GetRule 현재 규칙 반환
func (r *RuleRow) GetRule() *model.FirewallRule {
	return r.rule
}

// SetRule 규칙 설정 및 UI 업데이트
func (r *RuleRow) SetRule(rule *model.FirewallRule) {
	r.rule = rule
	r.syncFromRule()
}

// Content UI 컨테이너 반환
func (r *RuleRow) Content() *fyne.Container {
	return r.content
}
