package widget

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/internal/cache"
	"fyne.io/fyne/test"
	"fyne.io/fyne/theme"
)

func TestEntry_Cursor(t *testing.T) {
	entry := NewEntry()
	content := cache.Renderer(entry).(*entryRenderer).scroll.Content.(*entryContent)
	assert.Equal(t, desktop.TextCursor, content.Cursor())
}

func TestEntry_DoubleTapped(t *testing.T) {
	entry := NewEntry()
	entry.SetText("The quick brown fox\njumped    over the lazy dog\n")
	content := cache.Renderer(entry).(*entryRenderer).scroll.Content.(*entryContent)

	// select the word 'quick'
	ev := getClickPosition("The qui", 0)
	content.Tapped(ev)
	content.DoubleTapped(ev)
	assert.Equal(t, "quick", entry.SelectedText())

	// select the whitespace after 'quick'
	ev = getClickPosition("The quick", 0)
	// add half a ' ' character
	ev.Position.X += fyne.MeasureText(" ", theme.TextSize(), fyne.TextStyle{}).Width / 2
	content.Tapped(ev)
	content.DoubleTapped(ev)
	assert.Equal(t, " ", entry.SelectedText())

	// select all whitespace after 'jumped'
	ev = getClickPosition("jumped  ", 1)
	content.Tapped(ev)
	content.DoubleTapped(ev)
	assert.Equal(t, "    ", entry.SelectedText())
}

func TestEntry_DoubleTapped_AfterCol(t *testing.T) {
	entry := NewEntry()
	entry.SetText("A\nB\n")
	content := cache.Renderer(entry).(*entryRenderer).scroll.Content.(*entryContent)

	window := test.NewWindow(entry)
	defer window.Close()
	window.SetPadded(false)
	window.Resize(entry.MinSize())
	entry.Resize(entry.MinSize())
	c := window.Canvas()

	test.Tap(content)
	assert.Equal(t, entry, c.Focused())

	testCharSize := theme.TextSize()
	pos := fyne.NewPos(testCharSize, testCharSize*4) // tap below rows
	ev := &fyne.PointEvent{Position: pos}
	content.Tapped(ev)
	content.DoubleTapped(ev)

	assert.Equal(t, "", entry.SelectedText())
}

func TestEntry_DragSelect(t *testing.T) {
	entry := NewEntry()
	entry.SetText("The quick brown fox jumped\nover the lazy dog\nThe quick\nbrown fox\njumped over the lazy dog\n")
	content := cache.Renderer(entry).(*entryRenderer).scroll.Content.(*entryContent)

	// get position after the letter 'e' on the second row
	ev1 := getClickPosition("ove", 1)
	// get position after the letter 'z' on the second row
	ev2 := getClickPosition("over the laz", 1)
	// add a couple of pixels, this is currently a workaround for weird mouse to column logic on text with kerning
	ev2.Position.X += 2

	// mouse down and drag from 'r' to 'z'
	me := &desktop.MouseEvent{PointEvent: *ev1, Button: desktop.MouseButtonPrimary}
	content.MouseDown(me)
	for ; ev1.Position.X < ev2.Position.X; ev1.Position.X++ {
		de := &fyne.DragEvent{PointEvent: *ev1, Dragged: fyne.NewDelta(1, 0)}
		content.Dragged(de)
	}
	me = &desktop.MouseEvent{PointEvent: *ev1, Button: desktop.MouseButtonPrimary}
	content.MouseUp(me)

	assert.Equal(t, "r the laz", entry.SelectedText())
}

func TestEntry_ExpandSelectionForDoubleTap(t *testing.T) {
	str := []rune(" fish 日本語日  \t  test 本日本 moose  \t")

	// select invalid (before start)
	start, end := getTextWhitespaceRegion(str, -1)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	// select whitespace at the end of text
	start, end = getTextWhitespaceRegion(str, len(str))
	assert.Equal(t, 29, start)
	assert.Equal(t, 32, end)
	start, end = getTextWhitespaceRegion(str, len(str)+100)
	assert.Equal(t, 29, start)
	assert.Equal(t, 32, end)

	// select the whitespace
	start, end = getTextWhitespaceRegion(str, 0)
	assert.Equal(t, 0, start)
	assert.Equal(t, 1, end)

	// select "fish"
	start, end = getTextWhitespaceRegion(str, 1)
	assert.Equal(t, 1, start)
	assert.Equal(t, 5, end)
	start, end = getTextWhitespaceRegion(str, 4)
	assert.Equal(t, 1, start)
	assert.Equal(t, 5, end)

	// select "日本語日"
	start, end = getTextWhitespaceRegion(str, 6)
	assert.Equal(t, 6, start)
	assert.Equal(t, 10, end)
	start, end = getTextWhitespaceRegion(str, 9)
	assert.Equal(t, 6, start)
	assert.Equal(t, 10, end)

	// select "  \t  "
	start, end = getTextWhitespaceRegion(str, 10)
	assert.Equal(t, 10, start)
	assert.Equal(t, 15, end)

	// select "  \t"
	start, end = getTextWhitespaceRegion(str, 30)
	assert.Equal(t, 29, start)
	assert.Equal(t, len(str), end)
}

func TestEntry_ExpandSelectionWithWordSeparators(t *testing.T) {
	// select "is_a"
	str := []rune("This-is_a-test")
	start, end := getTextWhitespaceRegion(str, 6)
	assert.Equal(t, 5, start)
	assert.Equal(t, 9, end)
}

func TestEntry_EraseSelection(t *testing.T) {
	// Selects "sti" on line 2 of a new multiline
	// T e s t i n g
	// T e[s t i]n g
	// T e s t i n g
	e := NewMultiLineEntry()
	e.SetText("Testing\nTesting\nTesting")
	e.CursorRow = 1
	e.CursorColumn = 2
	var keyDown = func(key *fyne.KeyEvent) {
		e.KeyDown(key)
		e.TypedKey(key)
	}
	var keyPress = func(key *fyne.KeyEvent) {
		keyDown(key)
		e.KeyUp(key)
	}
	keyDown(&fyne.KeyEvent{Name: desktop.KeyShiftLeft})
	keyPress(&fyne.KeyEvent{Name: fyne.KeyRight})
	keyPress(&fyne.KeyEvent{Name: fyne.KeyRight})
	keyPress(&fyne.KeyEvent{Name: fyne.KeyRight})

	e.eraseSelection()
	e.updateText(e.textProvider().String())
	assert.Equal(t, "Testing\nTeng\nTesting", e.Text)
	a, b := e.selection()
	assert.Equal(t, -1, a)
	assert.Equal(t, -1, b)
}

func TestEntry_MouseClickAndDragOutsideText(t *testing.T) {
	entry := NewEntry()
	entry.SetText("A\nB\n")
	content := cache.Renderer(entry).(*entryRenderer).scroll.Content.(*entryContent)

	testCharSize := theme.TextSize()
	pos := fyne.NewPos(testCharSize, testCharSize*4) // tap below rows
	ev := &fyne.PointEvent{Position: pos}

	me := &desktop.MouseEvent{PointEvent: *ev, Button: desktop.MouseButtonPrimary}
	content.MouseDown(me)
	de := &fyne.DragEvent{PointEvent: *ev, Dragged: fyne.NewDelta(1, 0)}
	content.Dragged(de)
	content.MouseUp(me)
	assert.False(t, entry.selecting)
}

func TestEntry_MouseDownOnSelect(t *testing.T) {
	entry := NewEntry()
	entry.SetText("Ahnj\nBuki\n")
	entry.TypedShortcut(&fyne.ShortcutSelectAll{})
	content := cache.Renderer(entry).(*entryRenderer).scroll.Content.(*entryContent)

	testCharSize := theme.TextSize()
	pos := fyne.NewPos(testCharSize, testCharSize*4) // tap below rows
	ev := &fyne.PointEvent{Position: pos}

	me := &desktop.MouseEvent{PointEvent: *ev, Button: desktop.MouseButtonSecondary}
	content.MouseDown(me)
	content.MouseUp(me)

	assert.Equal(t, entry.SelectedText(), "Ahnj\nBuki\n")

	me = &desktop.MouseEvent{PointEvent: *ev, Button: desktop.MouseButtonPrimary}
	content.MouseDown(me)
	content.MouseUp(me)

	assert.Equal(t, entry.SelectedText(), "")
}

func TestEntry_PasteFromClipboard(t *testing.T) {
	entry := NewEntry()

	w := test.NewApp().NewWindow("")
	w.SetContent(entry)

	testContent := "test"

	clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	clipboard.SetContent(testContent)

	entry.pasteFromClipboard(clipboard)

	assert.Equal(t, entry.Text, testContent)
}

func getClickPosition(str string, row int) *fyne.PointEvent {
	x := fyne.MeasureText(str, theme.TextSize(), fyne.TextStyle{}).Width + theme.Padding()

	rowHeight := fyne.MeasureText("M", theme.TextSize(), fyne.TextStyle{}).Height
	y := theme.Padding() + float32(row)*rowHeight + rowHeight/2

	pos := fyne.NewPos(x, y)
	return &fyne.PointEvent{Position: pos}
}
