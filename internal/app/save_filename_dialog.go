package app

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
)

const saveFilenameExtension = ".xsp"

type saveFilenameDialog struct {
	Open   bool
	Text   string
	Cursor int
}

func (g *Game) openSaveFilenameDialog() {
	g.pathEntryCommand = pathEntryLabel("save")
	g.saveDialog = saveFilenameDialog{
		Open:   true,
		Text:   saveFilenameExtension,
		Cursor: 0,
	}
}

func (g *Game) SaveFilenameDialogOpen() bool {
	return g.saveDialog.Open
}

func (g *Game) SaveFilenameText() string {
	return g.saveDialog.Text
}

func (g *Game) SaveFilenameCursor() int {
	return g.saveDialog.Cursor
}

func (g *Game) EnterSaveFilenamePrefix(prefix string) {
	g.insertSaveFilenameText(prefix)
}

func (g *Game) insertSaveFilenameText(text string) {
	cursor := clampInt(g.saveDialog.Cursor, 0, len(g.saveDialog.Text))
	g.saveDialog.Text = g.saveDialog.Text[:cursor] + text + g.saveDialog.Text[cursor:]
	g.saveDialog.Cursor = cursor + len(text)
}

func (g *Game) deleteSaveFilenameCharacter() {
	if g.saveDialog.Cursor <= 0 || len(g.saveDialog.Text) == 0 {
		return
	}
	cursor := clampInt(g.saveDialog.Cursor, 0, len(g.saveDialog.Text))
	g.saveDialog.Text = g.saveDialog.Text[:cursor-1] + g.saveDialog.Text[cursor:]
	g.saveDialog.Cursor = cursor - 1
}

func (g *Game) clickSaveFilenameDialog(x int, y int) {
	point := image.Pt(x, y)
	if !point.In(saveFilenameDialogRect()) {
		g.saveDialog.Open = false
		return
	}
	if point.In(g.saveFilenameDialogOKRect()) {
		_ = g.SubmitSaveFilenameDialog()
	}
}

func (g *Game) SubmitSaveFilenameDialog() error {
	path, err := saveFilenamePath(g.saveDialog.Text)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(g.SaveXSP()), 0o600); err != nil {
		return err
	}
	g.currentFilePath = path
	g.saveDialog.Open = false
	return nil
}

func saveFilenamePath(input string) (string, error) {
	name := filepath.Base(strings.TrimSpace(input))
	name = strings.TrimSuffix(name, saveFilenameExtension)
	if name == "" {
		return "", fmt.Errorf("save filename is empty")
	}
	return filepath.Join("saves", name+saveFilenameExtension), nil
}

func (g *Game) CurrentFilePath() string {
	return g.currentFilePath
}

func saveFilenameDialogRect() image.Rectangle {
	x := screenWidth/2 - valueDialogWidth/2
	y := screenHeight/2 - valueDialogHeight/2
	return image.Rect(x, y, x+valueDialogWidth, y+valueDialogHeight)
}

func (g *Game) saveFilenameTextRect() image.Rectangle {
	rect := saveFilenameDialogRect()
	return image.Rect(rect.Min.X+12, rect.Min.Y+42, rect.Max.X-12, rect.Min.Y+66)
}

func (g *Game) saveFilenameDialogOKRect() image.Rectangle {
	rect := saveFilenameDialogRect()
	return image.Rect(rect.Max.X-64, rect.Max.Y-34, rect.Max.X-12, rect.Max.Y-12)
}
