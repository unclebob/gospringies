package app

import "image"

const (
	contextMenuWidth     = 120
	contextMenuTitleRows = 1
	contextMenuRowHeight = 24
	massMenuWidth        = contextMenuWidth
	massMenuTitleRows    = contextMenuTitleRows
	massMenuRowHeight    = contextMenuRowHeight
	springMenuWidth      = contextMenuWidth
	springMenuTitleRows  = contextMenuTitleRows
	springMenuRowHeight  = contextMenuRowHeight
)

type contextMenuItem struct {
	Label  string
	Action func()
}

func clickContextMenu(x int, y int, rect image.Rectangle, items []contextMenuItem, close func()) {
	point := image.Pt(x, y)
	if !point.In(rect) {
		close()
		return
	}
	for i, item := range items {
		if point.In(contextMenuRowRect(rect, i)) {
			item.Action()
			close()
			return
		}
	}
}

func contextMenuLabels(items []contextMenuItem) []string {
	var labels []string
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return labels
}

func selectContextMenuItem(items []contextMenuItem, label string, close func()) bool {
	for _, item := range items {
		if item.Label == label {
			item.Action()
			close()
			return true
		}
	}
	return false
}

func contextMenuRect(x int, y int, itemCount int) image.Rectangle {
	rows := contextMenuTitleRows + itemCount
	height := rows * contextMenuRowHeight
	left := clampInt(x, 0, screenWidth-contextMenuWidth)
	top := clampInt(y, 0, screenHeight-height)
	return image.Rect(left, top, left+contextMenuWidth, top+height)
}

func contextMenuRowRect(rect image.Rectangle, index int) image.Rectangle {
	top := rect.Min.Y + contextMenuTitleRows*contextMenuRowHeight + index*contextMenuRowHeight
	return image.Rect(rect.Min.X, top, rect.Max.X, top+contextMenuRowHeight)
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-21T11:40:22-05:00","module_hash":"8b0ede2475fee8974faf2dc3ea2dbb2d6da902c7dfefd555e4088e052ad2a719","functions":[{"id":"func/clickContextMenu","name":"clickContextMenu","line":22,"end_line":35,"hash":"5d12f169871a0f66e1e307585b9f2265cb0db8665676f6a07fee0144465c5500"},{"id":"func/contextMenuLabels","name":"contextMenuLabels","line":37,"end_line":43,"hash":"7fa54d22e9071d91eaa4f8e39470dbd2d9fcc33e9fdf635a6f83e116e568215f"},{"id":"func/selectContextMenuItem","name":"selectContextMenuItem","line":45,"end_line":54,"hash":"eea8e3721b30f429ba3969a6a2069305e628696f9141baa76680c87778b055d1"},{"id":"func/contextMenuRect","name":"contextMenuRect","line":56,"end_line":62,"hash":"2a18853e28684b07131810b9d11c3fc3a7bb130637cc96795468b2973b71bdff"},{"id":"func/contextMenuRowRect","name":"contextMenuRowRect","line":64,"end_line":67,"hash":"45ba9111b92453881485bfe055ad9d9c280e39c4adcedfc191af76020f8b30a3"}]}
// mutate4go-manifest-end
