package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// 라벨과 입력 필드를 함께 표시합니다.
func NewLabeledEntry(label, placeholder string, onChange func(string)) *fyne.Container {
	lbl := widget.NewLabel(label)
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)

	if onChange != nil {
		entry.OnChanged = onChange
	}

	return container.NewBorder(nil, nil, lbl, nil, entry)
}

// 라벨과 비밀번호 입력 필드를 함께 표시합니다.
func NewLabeledPassword(label, placeholder string, onChange func(string)) *fyne.Container {
	lbl := widget.NewLabel(label)
	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder(placeholder)

	if onChange != nil {
		entry.OnChanged = onChange
	}

	return container.NewBorder(nil, nil, lbl, nil, entry)
}

// 라벨과 드롭다운을 함께 표시합니다.
func NewLabeledSelect(label string, options []string, onChange func(string)) *fyne.Container {
	lbl := widget.NewLabel(label)
	sel := widget.NewSelect(options, onChange)

	return container.NewBorder(nil, nil, lbl, nil, sel)
}

// 라벨과 여러 줄 입력 필드를 함께 표시합니다.
func NewLabeledMultiLineEntry(label string, onChange func(string)) *fyne.Container {
	lbl := widget.NewLabel(label)
	entry := widget.NewMultiLineEntry()

	if onChange != nil {
		entry.OnChanged = onChange
	}

	return container.NewBorder(lbl, nil, nil, nil, entry)
}
