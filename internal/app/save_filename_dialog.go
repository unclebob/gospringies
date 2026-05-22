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
	g.document.pathEntryCommand = pathEntryLabel("save")
	g.overlays.save = saveFilenameDialog{
		Open:   true,
		Text:   saveFilenameExtension,
		Cursor: 0,
	}
}

func (g *Game) SaveFilenameDialogOpen() bool {
	return g.overlays.save.Open
}

func (g *Game) SaveFilenameText() string {
	return g.overlays.save.Text
}

func (g *Game) SaveFilenameCursor() int {
	return g.overlays.save.Cursor
}

func (g *Game) EnterSaveFilenamePrefix(prefix string) {
	g.insertSaveFilenameText(prefix)
}

func (g *Game) insertSaveFilenameText(text string) {
	cursor := clampInt(g.overlays.save.Cursor, 0, len(g.overlays.save.Text))
	g.overlays.save.Text = g.overlays.save.Text[:cursor] + text + g.overlays.save.Text[cursor:]
	g.overlays.save.Cursor = cursor + len(text)
}

func (g *Game) deleteSaveFilenameCharacter() {
	if g.overlays.save.Cursor <= 0 || len(g.overlays.save.Text) == 0 {
		return
	}
	cursor := clampInt(g.overlays.save.Cursor, 0, len(g.overlays.save.Text))
	g.overlays.save.Text = g.overlays.save.Text[:cursor-1] + g.overlays.save.Text[cursor:]
	g.overlays.save.Cursor = cursor - 1
}

func (g *Game) clickSaveFilenameDialog(x int, y int) {
	point := image.Pt(x, y)
	if !point.In(saveFilenameDialogRect()) {
		g.overlays.save.Open = false
		return
	}
	if point.In(g.saveFilenameDialogOKRect()) {
		_ = g.SubmitSaveFilenameDialog()
	}
}

func (g *Game) SubmitSaveFilenameDialog() error {
	path, err := saveFilenamePath(g.overlays.save.Text)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(g.SaveXSP()), 0o600); err != nil {
		return err
	}
	g.document.currentFilePath = path
	g.overlays.save.Open = false
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
	return g.document.currentFilePath
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:11:49-05:00","module_hash":"3f984856101e002340522fe08fca239f31d0b3287ab18b0f34b7c4f5cf745fd2","functions":[{"id":"func/Game.openSaveFilenameDialog","name":"Game.openSaveFilenameDialog","line":19,"end_line":26,"hash":"1a840857537a97d070443e635cef89a63665e7b7982f0a9f31c14ed82eb4eee1"},{"id":"func/Game.SaveFilenameDialogOpen","name":"Game.SaveFilenameDialogOpen","line":28,"end_line":30,"hash":"19517f003cdff49b8a8c625d86559e561b3381aca446e9f102aea74a747750a1"},{"id":"func/Game.SaveFilenameText","name":"Game.SaveFilenameText","line":32,"end_line":34,"hash":"3cba88273a7ec614c46f177dbdc501d5e514a4f4549dfdd8ce9c38415bcd695f"},{"id":"func/Game.SaveFilenameCursor","name":"Game.SaveFilenameCursor","line":36,"end_line":38,"hash":"a3e3209dfea2fcaa6392114d56ae52afa1c476e9a6eb997779cfa4d657e32198"},{"id":"func/Game.EnterSaveFilenamePrefix","name":"Game.EnterSaveFilenamePrefix","line":40,"end_line":42,"hash":"2ef8dd618c5aa0be80a5cf860f1ef35bebc032d03cb36702042a96b16bae5d69"},{"id":"func/Game.insertSaveFilenameText","name":"Game.insertSaveFilenameText","line":44,"end_line":48,"hash":"73393d173ea20090496f169f6b02f550102aa68c0071b8b3f851ddb16cb4fa8c"},{"id":"func/Game.deleteSaveFilenameCharacter","name":"Game.deleteSaveFilenameCharacter","line":50,"end_line":57,"hash":"75645ab265b2bcc2c4dbfd99eab5070160f0fb58adb567f7ce8a0791f6ef0029"},{"id":"func/Game.clickSaveFilenameDialog","name":"Game.clickSaveFilenameDialog","line":59,"end_line":68,"hash":"f8cb2fce3aaa0bf6b4c5b145a3dbecb613e95ade24eedcd6235411bb34a4d99c"},{"id":"func/Game.SubmitSaveFilenameDialog","name":"Game.SubmitSaveFilenameDialog","line":70,"end_line":84,"hash":"5ffe17d0828b5fd07e4f097df3e0c78e199f46f3493ee81f85938c75e2aa64e4"},{"id":"func/saveFilenamePath","name":"saveFilenamePath","line":86,"end_line":93,"hash":"ff6b63277a6932c1ecc5ea2405561e554f864e3dfc918b02d4fe019ddf9f27fc"},{"id":"func/Game.CurrentFilePath","name":"Game.CurrentFilePath","line":95,"end_line":97,"hash":"e1938656214c4848955a5871e65cb88d7de0d535982c5ec020f6076ae4764d05"},{"id":"func/saveFilenameDialogRect","name":"saveFilenameDialogRect","line":99,"end_line":103,"hash":"5b337f75d4db87d22232e796998e95b6b56fe4b3c8a8502dd0fc598555608ebc"},{"id":"func/Game.saveFilenameTextRect","name":"Game.saveFilenameTextRect","line":105,"end_line":108,"hash":"e0db63d9fbf2a6e5727ffb6dc879f89186a65f1d96ed60352c978b2e4dbb3733"},{"id":"func/Game.saveFilenameDialogOKRect","name":"Game.saveFilenameDialogOKRect","line":110,"end_line":113,"hash":"2e24bf879d33b88e92875fb6b1d14cadeaeecf9072fd87464b607b9b4e971135"}]}
// mutate4go-manifest-end
