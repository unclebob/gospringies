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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T15:32:43-05:00","module_hash":"70e3c46c2469f317f1b0c18ba46d02912bd97a7a54cf8c014fc5ce3419152582","functions":[{"id":"func/Game.openSaveFilenameDialog","name":"Game.openSaveFilenameDialog","line":19,"end_line":26,"hash":"0d85296ed9014a1dd9188d00ec6253d6371a2065d9e269102b4215db23f6f8b7"},{"id":"func/Game.SaveFilenameDialogOpen","name":"Game.SaveFilenameDialogOpen","line":28,"end_line":30,"hash":"08614f904911cd3983402f450e155ae372532314bf152c36f0bb42e1ca9a3feb"},{"id":"func/Game.SaveFilenameText","name":"Game.SaveFilenameText","line":32,"end_line":34,"hash":"7f2a2337549947d2bc90b8f5c2948fd8dd4c70ca1de89e9cea6a570e03fd93a6"},{"id":"func/Game.SaveFilenameCursor","name":"Game.SaveFilenameCursor","line":36,"end_line":38,"hash":"faca3625c053a133bd6bc7e231898ec415fbbef0183a049b56252b73c991511c"},{"id":"func/Game.EnterSaveFilenamePrefix","name":"Game.EnterSaveFilenamePrefix","line":40,"end_line":42,"hash":"2ef8dd618c5aa0be80a5cf860f1ef35bebc032d03cb36702042a96b16bae5d69"},{"id":"func/Game.insertSaveFilenameText","name":"Game.insertSaveFilenameText","line":44,"end_line":48,"hash":"8af6da633bec9e7c799d35c4f9941ea3b077d9f950d120b5db45603153567473"},{"id":"func/Game.deleteSaveFilenameCharacter","name":"Game.deleteSaveFilenameCharacter","line":50,"end_line":57,"hash":"3d08ceb9de592b80e83899d149ae90d1c57cadf3158df5499d4fb7d3b39e12b0"},{"id":"func/Game.clickSaveFilenameDialog","name":"Game.clickSaveFilenameDialog","line":59,"end_line":68,"hash":"205bc50bf8f1e4830b38d7725368b314239059e338c7b40ebe6dc71ea5e70408"},{"id":"func/Game.SubmitSaveFilenameDialog","name":"Game.SubmitSaveFilenameDialog","line":70,"end_line":84,"hash":"b344e09ca738cb60523166fe9aca168737bda97217bdba53ea7d7406c2e7cf8c"},{"id":"func/saveFilenamePath","name":"saveFilenamePath","line":86,"end_line":93,"hash":"ff6b63277a6932c1ecc5ea2405561e554f864e3dfc918b02d4fe019ddf9f27fc"},{"id":"func/Game.CurrentFilePath","name":"Game.CurrentFilePath","line":95,"end_line":97,"hash":"3c5aa42c626811367195ed88190077f2d60a1f77e55d3846b2e699771645074d"},{"id":"func/saveFilenameDialogRect","name":"saveFilenameDialogRect","line":99,"end_line":103,"hash":"5b337f75d4db87d22232e796998e95b6b56fe4b3c8a8502dd0fc598555608ebc"},{"id":"func/Game.saveFilenameTextRect","name":"Game.saveFilenameTextRect","line":105,"end_line":108,"hash":"e0db63d9fbf2a6e5727ffb6dc879f89186a65f1d96ed60352c978b2e4dbb3733"},{"id":"func/Game.saveFilenameDialogOKRect","name":"Game.saveFilenameDialogOKRect","line":110,"end_line":113,"hash":"2e24bf879d33b88e92875fb6b1d14cadeaeecf9072fd87464b607b9b4e971135"}]}
// mutate4go-manifest-end
