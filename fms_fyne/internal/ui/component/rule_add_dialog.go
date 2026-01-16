package component

import (
	"fms/internal/model"
	"fms/internal/themes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// RuleAddDialog 규칙 추가 다이얼로그
type RuleAddDialog struct {
	window fyne.Window
	onAdd  func(*model.FirewallRule)

	// 내부 폼
	generalForm    *RuleForm
	blackWhiteForm *BlackWhiteForm
	tabs           *container.AppTabs

	// 다이얼로그 팝업
	popup *widget.PopUp
}

// NewRuleAddDialog 새 규칙 추가 다이얼로그 생성
func NewRuleAddDialog(window fyne.Window, onAdd func(*model.FirewallRule)) *RuleAddDialog {
	d := &RuleAddDialog{
		window: window,
		onAdd:  onAdd,
	}
	d.createUI()
	return d
}

// createUI 다이얼로그 UI 생성
func (d *RuleAddDialog) createUI() {
	// 폼 생성 (헤더/버튼 포함된 기존 폼 사용)
	d.generalForm = NewRuleForm(d.onAdd)
	d.blackWhiteForm = NewBlackWhiteForm(d.onAdd)

	// 탭 구성 (내부 패딩 10)
	d.tabs = container.NewAppTabs(
		container.NewTabItem("일반 규칙", container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), d.generalForm.Content())),
		container.NewTabItem("Black/White", container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), d.blackWhiteForm.Content())),
	)

	// 닫기 버튼 (배경 없이 텍스트만)
	closeBtn := NewCustomButton("X", nil, themes.Colors["black"], nil, func() {
		d.Hide()
	})

	// 헤더
	header := container.NewBorder(
		nil, nil,
		widget.NewLabelWithStyle("규칙 추가", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		closeBtn,
	)

	// 추가 버튼 (오른쪽 하단, 외부 패딩: 하단 10, 우측 10)
	addBtn := NewCustomButton("+ 추가", nil, nil, themes.Colors["darkgray"], func() {
		// 현재 선택된 탭에 따라 해당 폼의 SubmitRule 호출
		var success bool
		selectedTab := d.tabs.Selected()
		if selectedTab != nil {
			switch selectedTab.Text {
			case "일반 규칙":
				success = d.generalForm.SubmitRule()
			case "Black/White":
				success = d.blackWhiteForm.SubmitRule()
			}
		}
		// 성공 시 다이얼로그 닫기
		if success {
			d.Hide()
		}
	})
	// 오른쪽 정렬 + 외부 패딩 (상, 하, 좌, 우)
	addBtnContainer := container.New(layout.NewCustomPaddedLayout(0, 10, 0, 10),
		container.NewHBox(layout.NewSpacer(), addBtn),
	)

	// 전체 레이아웃
	content := container.NewBorder(
		header,
		addBtnContainer, nil, nil,
		d.tabs,
	)

	// 고정 크기 컨테이너
	sizedContent := container.NewGridWrap(fyne.NewSize(700, 400), content)

	// 팝업 생성 (처음에는 숨김)
	d.popup = widget.NewModalPopUp(sizedContent, d.window.Canvas())
}

// Show 다이얼로그 표시
func (d *RuleAddDialog) Show() {
	// 폼 초기화
	d.generalForm.Reset()
	d.blackWhiteForm.Reset()

	// 첫 번째 탭 선택
	if len(d.tabs.Items) > 0 {
		d.tabs.SelectIndex(0)
	}

	d.popup.Show()
}

// Hide 다이얼로그 숨김
func (d *RuleAddDialog) Hide() {
	d.popup.Hide()
}

// ResetTabs 탭 초기화 (첫 번째 탭 선택)
func (d *RuleAddDialog) ResetTabs() {
	if len(d.tabs.Items) > 0 {
		d.tabs.SelectIndex(0)
	}
}
