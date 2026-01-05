package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// 확인/취소 다이얼로그를 표시합니다.
func ShowConfirmDialog(parent fyne.Window, title, message string, onConfirm func()) {
	dialog.ShowConfirm(title, message, func(confirmed bool) {
		if confirmed && onConfirm != nil {
			onConfirm()
		}
	}, parent)
}

// 알림 다이얼로그를 표시합니다.
func ShowAlertDialog(parent fyne.Window, title, message string) {
	dialog.ShowInformation(title, message, parent)
}

// 에러 다이얼로그를 표시합니다.
func ShowErrorDialog(parent fyne.Window, err error) {
	dialog.ShowError(err, parent)
}

// 진행률 다이얼로그를 표시합니다.
func ShowProgressDialog(parent fyne.Window, title string) *dialog.CustomDialog {
	progress := widget.NewProgressBarInfinite()
	return dialog.NewCustom(title, "취소", progress, parent)
}

// 입력 다이얼로그를 표시합니다.
func ShowInputDialog(parent fyne.Window, title, placeholder string, onSubmit func(string)) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)

	dialog.ShowForm(title, "확인", "취소", []*widget.FormItem{
		widget.NewFormItem("", entry),
	}, func(confirmed bool) {
		if confirmed && onSubmit != nil {
			onSubmit(entry.Text)
		}
	}, parent)
}
