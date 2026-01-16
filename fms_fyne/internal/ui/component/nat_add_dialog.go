package component

import (
	"fms/internal/model"
	"fms/internal/themes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// NATAddDialog NAT 규칙 추가 다이얼로그
type NATAddDialog struct {
	window fyne.Window
	onAdd  func(*model.NATRule)

	// 내부 폼
	dnatForm *DNATForm
	snatForm *SNATForm
	tabs     *container.AppTabs

	// 다이얼로그 팝업
	popup *widget.PopUp
}

// NewNATAddDialog 새 NAT 규칙 추가 다이얼로그 생성
func NewNATAddDialog(window fyne.Window, onAdd func(*model.NATRule)) *NATAddDialog {
	d := &NATAddDialog{
		window: window,
		onAdd:  onAdd,
	}
	d.createUI()
	return d
}

// createUI 다이얼로그 UI 생성
func (d *NATAddDialog) createUI() {
	// 폼 생성 (헤더/버튼 포함된 기존 폼 사용)
	d.dnatForm = NewDNATForm(d.onAdd)
	d.snatForm = NewSNATForm(d.onAdd)

	// 탭 구성 (내부 패딩 10)
	d.tabs = container.NewAppTabs(
		container.NewTabItem("DNAT (포트 포워딩)", container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), d.dnatForm.Content())),
		container.NewTabItem("SNAT/MASQUERADE", container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), d.snatForm.Content())),
	)

	// 닫기 버튼 (배경 없이 텍스트만)
	closeBtn := NewCustomButton("X", nil, themes.Colors["black"], nil, func() {
		d.Hide()
	})

	// 헤더
	header := container.NewBorder(
		nil, nil,
		widget.NewLabelWithStyle("NAT 규칙 추가", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		closeBtn,
	)

	// 추가 버튼 (오른쪽 하단, 외부 패딩: 하단 10, 우측 10)
	addBtn := NewCustomButton("+ 추가", nil, nil, themes.Colors["darkgray"], func() {
		// 현재 선택된 탭에 따라 해당 폼의 SubmitRule 호출
		var success bool
		selectedTab := d.tabs.Selected()
		if selectedTab != nil {
			switch selectedTab.Text {
			case "DNAT (포트 포워딩)":
				success = d.dnatForm.SubmitRule()
			case "SNAT/MASQUERADE":
				success = d.snatForm.SubmitRule()
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
	sizedContent := container.NewGridWrap(fyne.NewSize(700, 350), content)

	// 팝업 생성 (처음에는 숨김)
	d.popup = widget.NewModalPopUp(sizedContent, d.window.Canvas())
}

// Show 다이얼로그 표시
func (d *NATAddDialog) Show() {
	// 폼 초기화
	d.dnatForm.Reset()
	d.snatForm.Reset()

	// 첫 번째 탭 선택
	if len(d.tabs.Items) > 0 {
		d.tabs.SelectIndex(0)
	}

	d.popup.Show()
}

// Hide 다이얼로그 숨김
func (d *NATAddDialog) Hide() {
	d.popup.Hide()
}

// ResetTabs 탭 초기화 (첫 번째 탭 선택)
func (d *NATAddDialog) ResetTabs() {
	if len(d.tabs.Items) > 0 {
		d.tabs.SelectIndex(0)
	}
}
