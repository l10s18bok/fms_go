package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// 데이터 테이블을 생성합니다.
func NewDataTable(headers []string, data [][]string, onSelect func(row int)) *widget.Table {
	table := widget.NewTable(
		// 크기 함수
		func() (int, int) {
			return len(data) + 1, len(headers) // +1 for header
		},
		// 셀 생성 함수
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		// 셀 업데이트 함수
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row == 0 {
				// 헤더
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				// 데이터
				if id.Row-1 < len(data) && id.Col < len(data[id.Row-1]) {
					label.SetText(data[id.Row-1][id.Col])
				}
			}
		},
	)

	if onSelect != nil {
		table.OnSelected = func(id widget.TableCellID) {
			if id.Row > 0 { // 헤더 제외
				onSelect(id.Row - 1)
			}
		}
	}

	return table
}

// 테이블 열 너비를 설정합니다.
func SetColumnWidths(table *widget.Table, widths []float32) {
	for i, width := range widths {
		table.SetColumnWidth(i, width)
	}
}
