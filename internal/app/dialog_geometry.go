package app

import "image"

func centeredDialogRect(width int, height int) image.Rectangle {
	x := screenWidth/2 - width/2
	y := screenHeight/2 - height/2
	return image.Rect(x, y, x+width, y+height)
}

func dialogTextRect(rect image.Rectangle) image.Rectangle {
	return image.Rect(rect.Min.X+12, rect.Min.Y+42, rect.Max.X-12, rect.Min.Y+66)
}

func dialogOKRect(rect image.Rectangle) image.Rectangle {
	return image.Rect(rect.Max.X-64, rect.Max.Y-34, rect.Max.X-12, rect.Max.Y-12)
}
