package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// 제목이 있는 카드 컨테이너를 생성합니다.
func NewCard(title string, content fyne.CanvasObject) *widget.Card {
	return widget.NewCard(title, "", content)
}

// 제목과 부제목이 있는 카드를 생성합니다.
func NewCardWithSubtitle(title, subtitle string, content fyne.CanvasObject) *widget.Card {
	return widget.NewCard(title, subtitle, content)
}

// 제목과 액션 버튼이 있는 카드를 생성합니다.
func NewCardWithActions(title string, content fyne.CanvasObject, actions ...fyne.CanvasObject) *fyne.Container {
	header := container.NewBorder(
		nil, nil, widget.NewLabel(title), container.NewHBox(actions...),
	)

	return container.NewBorder(header, nil, nil, nil, content)
}
