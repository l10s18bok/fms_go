package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// 체크 가능한 아이템입니다.
type CheckableItem struct {
	ID      int
	Label   string
	Checked bool
}

// 체크박스가 있는 목록을 생성합니다.
func NewCheckableList(items []CheckableItem, onChange func(id int, checked bool)) *fyne.Container {
	vbox := container.NewVBox()

	for _, item := range items {
		itemCopy := item // 클로저용 복사
		check := widget.NewCheck(item.Label, func(checked bool) {
			if onChange != nil {
				onChange(itemCopy.ID, checked)
			}
		})
		check.Checked = item.Checked
		vbox.Add(check)
	}

	return vbox
}

// 목록의 모든 체크박스를 선택/해제합니다.
func SelectAllInList(list *fyne.Container, checked bool) {
	for _, obj := range list.Objects {
		if check, ok := obj.(*widget.Check); ok {
			check.SetChecked(checked)
		}
	}
}
