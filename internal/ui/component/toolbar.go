package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// 툴바 아이템입니다.
type ToolbarItem struct {
	Label string
	Icon  fyne.Resource
	OnTap func()
}

// 툴바를 생성합니다.
func NewToolbar(items ...ToolbarItem) *widget.Toolbar {
	toolbarItems := make([]widget.ToolbarItem, len(items))

	for i, item := range items {
		itemCopy := item
		if item.Icon != nil {
			toolbarItems[i] = widget.NewToolbarAction(item.Icon, itemCopy.OnTap)
		} else {
			toolbarItems[i] = widget.NewToolbarAction(nil, itemCopy.OnTap)
		}
	}

	return widget.NewToolbar(toolbarItems...)
}

// 스페이서가 포함된 툴바를 생성합니다.
func NewToolbarWithSpacer(left []ToolbarItem, right []ToolbarItem) *widget.Toolbar {
	items := []widget.ToolbarItem{}

	for _, item := range left {
		itemCopy := item
		items = append(items, widget.NewToolbarAction(item.Icon, itemCopy.OnTap))
	}

	items = append(items, widget.NewToolbarSpacer())

	for _, item := range right {
		itemCopy := item
		items = append(items, widget.NewToolbarAction(item.Icon, itemCopy.OnTap))
	}

	return widget.NewToolbar(items...)
}
