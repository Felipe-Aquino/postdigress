package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  //"strings"
  //"fmt"
)

type StatusMode byte

const (
  Show    StatusMode = iota
  Prompt
)

type Status struct {
  tv *tview.TextView

  text Text
  textBuffer Text
  startWith string

  cursor int
  mode StatusMode

  onEnter  func(string)
  onCancel func()

  history *DumbHistory
}

func NewStatus() *Status {
  s := &Status{
    tv: tview.NewTextView(),
    text: WrapLines(""),
    mode: Show,
    onEnter: func(s string) {},
    onCancel: func() {},
    history: NewDumbHistory(10),
    textBuffer: nil,
    startWith: "",
  }

	s.tv.
    SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true).
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      //fmt.Printf("name: %s\n", event.Name())
      if event.Key() == tcell.KeyRune {
        s.HandleKeyboard(event.Rune(), 0)
      } else {
        s.HandleKeyboard(rune(0), event.Key())
      }
      return nil
    })
  return s
}

func (s *Status) ChangeStartString(str string) {
  s.startWith = str
}

func (s *Status) SetMode(m StatusMode) {
  s.mode = m

  if m == Prompt {
    s.Clear()
    s.UpdateText()
    s.tv.Highlight("cursor")
  } else {
    s.tv.Highlight("")
  }
}

func (s *Status) Clear() {
  s.text.SetLine(0, "")
}

func (s *Status) SaveHistory() {
  es := NewEditorState(s.text.Clone(), s.cursor, 0)
  s.history.Push(es)
}

// TODO: Duplicated
func SInsertCursorTag(s string, at int) string {
  return s[:at] + "[\"cursor\"]" + string(s[at]) + "[\"\"]" + s[at + 1:] 
}

func (s *Status) SetEnterCb(cb func(string)) {
  s.onEnter = cb
}

func (s *Status) SetCancelCb(cb func()) {
  s.onCancel = cb
}

func (s *Status) MoveCursorRight() {
  if s.cursor < s.text.LineLen(0) {
    s.cursor++
  }
}

func (s *Status) MoveCursorLeft() {
  if s.cursor > 0 {
    s.cursor--
  }
}

func (s *Status) MoveCursorUp() {
  state := s.history.Undo()
  if state != nil {
    if s.textBuffer == nil {
      s.textBuffer = s.text
    }

    s.text, _, _ = state.Unpack()
  }

  s.cursor = s.text.LineLen(0)
}

func (s *Status) MoveCursorDown() {
  state := s.history.Redo()
  if state != nil {
    s.text, _, _ = state.Unpack()
  } else if s.textBuffer != nil {
    s.text = s.textBuffer
    s.textBuffer = nil
  }

  s.cursor = s.text.LineLen(0)
}

func (s *Status) HandleKeyboard(ch rune, key tcell.Key) {
  //fmt.Printf("- %V %V %V\n", ch, key, tcell.KeyBS)
  if s.mode == Prompt {
    switch key {
    case tcell.KeyESC:
      s.Clear()
      s.SetMode(Show)
      s.onCancel()
      return
    case tcell.KeyDown, tcell.KeyCtrlJ, tcell.KeyCtrlN:
      s.MoveCursorDown()
    case tcell.KeyUp, tcell.KeyCtrlK:
      s.MoveCursorUp()
    case tcell.KeyRight, tcell.KeyCtrlL, tcell.KeyCtrlP:
      s.MoveCursorRight()
    case tcell.KeyLeft, tcell.KeyCtrlH:
      s.MoveCursorLeft()
    case tcell.KeyBackspace2:
      if s.cursor > 0 {
        s.text = s.text.DeleteRange(0, s.cursor - 1, 0, s.cursor - 1)
        s.MoveCursorLeft()
      }
    case tcell.KeyCR:
      if s.text.Line(0) != "" {
        s.SaveHistory()
      }

      text := s.text.Line(0)
      s.Clear()

      s.onEnter(text)
      s.cursor = 0
    default:
      if key == 0 {
        s.history.RedoToLast()
        s.textBuffer = nil

        s.text = s.text.InsertAt(0, s.cursor, WrapLines(string(ch)))
        s.MoveCursorRight()
      }
    }

    s.UpdateText()
  }
}

func (s *Status) SetText(text string) {
  s.text.SetLine(0, text)
  s.UpdateText()
}

func (s *Status) UpdateText() {
  if s.mode == Prompt {
    cursor := s.cursor + len(s.startWith)

    text := SInsertCursorTag(s.startWith + s.text.Line(0) + " ", cursor)
    s.tv.SetText(text)
    s.tv.ScrollTo(cursor, 0)
  } else {
    s.tv.SetText(s.text.Line(0))
  }
}

func dddmain() {
	app := tview.NewApplication()

  command := NewCommand()

  result := tview.NewTextView()

  bar := NewStatus()
  bar.ChangeStartString(":")
  bar.SetMode(Prompt)

  bar.SetEnterCb(func(s string) {
    returned, err := command.Run(s)

    if err != nil {
      result.SetText(s + "\n" + err.Error())
    } else {
      result.SetText(s + "\n" + returned)
    }
  })

  bar.SetCancelCb(func() {

  })

  bar.tv.Highlight("cursor")

  flex := tview.NewFlex().
    SetDirection(tview.FlexRow).
    AddItem(bar.tv, 1, 1, true).
    AddItem(result, 0, 1, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}

}
