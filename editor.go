package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "strings"
  "fmt"
)

type HColor int

const (
  Red HColor = iota
  Pink
  Yellow
  White
  Violet
  Blue
  Green
  Orange
  Gray
  Wheat
  Turquoise
)

type Highlight struct {
  start, end int
  color HColor
}

type Mode rune

const (
  NORMAL Mode = 'N'
  INSERT Mode = 'I'
  VISUAL Mode = 'V'
)

type VisualSelect struct {
  start, size int
}

type Editor struct {
  tv *tview.TextView
  history *DumbHistory

  lines []string
  mode Mode

  cursorX, cursorY int

  tokenizer *Tokenizer
  highlights []Highlight
  fullText string

  onModeChanged func(Mode)
  onExecute func(string)

  selected VisualSelect

  buffCommand string // buffer that save a multi-letter command
  yankedBuffer string  // buffer use to save copied/deleted text
}

func NewEditor() *Editor {
  e := &Editor{
    tv: tview.NewTextView(),
    lines: []string{
      "-- just a commnet", 
      "select * from test;",
      "",
      "insert into test (name, value, created_at) values",
      "('fifth', 30, '2020-10-10 22:22:22Z')",
      "",
      "/* multiline comment */",
    },
    mode: NORMAL,
    onModeChanged: func(m Mode) {
    },
    onExecute: func(s string) {
    },
  }

  e.history = NewDumbHistory(5)

  e.tokenizer = NewTokenizer()
  e.fullText = ""

	e.tv.
    SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true).
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      //fmt.Printf("name: %s\n", event.Name())
      if event.Key() == tcell.KeyRune {
        e.HandleKeyboard(event.Rune(), 0)
      } else {
        e.HandleKeyboard(rune(0), event.Key())
      }
      return nil
    })

  //e.tv.Highlight("cursor")

  return e
}

func (e *Editor) SaveHistory() {
  es := NewEditorState(SClone(e.lines), e.cursorX, e.cursorY)
  e.history.Add(es)
}

func (e *Editor) AddLine(line string) {
  e.lines = append(e.lines, line)
}

func (e *Editor) AddLineAt(line string, at int) {
  if at >= 0 && at < len(e.lines) {
    e.lines = append(e.lines, "")
    copy(e.lines[at + 1:] , e.lines[at:])
    e.lines[at] = line
  } else {
    e.AddLine(line)
  }
}

func (e *Editor) NewLineAfter(row, col int) {
  if row >= 0 && row < len(e.lines) {
    currentLine := e.lines[row][:col]
    newLine := e.lines[row][col:]

    e.lines = append(e.lines, "")
    copy(e.lines[row + 1:] , e.lines[row:])
    e.lines[row] = currentLine
    e.lines[row + 1] = newLine
  } else {
    row = Max(0, len(e.lines) - 1)

    currentLine := e.lines[row][:col]
    newLine := e.lines[row][col:]

    e.lines[row] = currentLine 
    e.lines = append(e.lines, newLine)
  }

  e.fullText = ""
}

func (e *Editor) InsertYankedWordAfter(row, col int) {
  if row >= 0 && row < len(e.lines) {
    line := e.lines[row]

    if col <= 0 {
      e.lines[row] = ""
      if len(e.lines[row]) != 0 {
        e.lines[row] = " "
      }
      e.lines[row] += e.yankedBuffer + line
    } else if col < len(line) - 1 {
      e.lines[row] = line[:col + 1] + e.yankedBuffer + line[col + 1:]
    } else {
      e.lines[row] = line + e.yankedBuffer
    }

    e.fullText = ""
  }
}

func (e *Editor) InsertCharBefore(ch rune, row, col int) {
  if row >= 0 && row < len(e.lines) {
    line := e.lines[row]

    if col <= 0 {
      e.lines[row] = string(ch) + line
    } else if col < len(line) {
      e.lines[row] = line[:col] + string(ch) + line[col:]
    } else {
      e.lines[row] = line + string(ch)
    }

    e.fullText = ""
  }
}

func (e *Editor) DeleteCharBefore(row, col int) {
  if row >= 0 && row < len(e.lines) {
    line := e.lines[row]

    lineLen := len(line)
    if col > 0 && col < lineLen {
      e.lines[row] = line[:col - 1] + line[col:]

    } else if col >= lineLen && lineLen > 0 {
      e.lines[row] = line[:lineLen - 1]

    } else if col == 0 && row > 0 {
      e.lines[row - 1] = e.lines[row - 1] + e.lines[row]
      e.lines = append(e.lines[:row], e.lines[row + 1:]...)
    }
    e.fullText = ""
  }
}

func (e *Editor) MoveCursorUp() {
  if e.cursorY > 0 {
    e.cursorY--

    lineLen := len(e.lines[e.cursorY])
    if e.cursorX > lineLen - 1 {
      e.cursorX = Max(0, lineLen - 1)
    }
  }
}

func (e *Editor) MoveCursorDown() {
  if e.cursorY < len(e.lines) - 1 {
    e.cursorY++

    lineLen := len(e.lines[e.cursorY])
    if e.cursorX > lineLen - 1 {
      e.cursorX = Max(0, lineLen - 1)
    }
  }
}

func (e *Editor) MoveCursorRight() {
  if e.cursorX < len(e.lines[e.cursorY]) {
    e.cursorX++
  }
}

func (e *Editor) MoveCursorLeft() {
  if e.cursorX > 0 {
    e.cursorX--
  }
}

func (e *Editor) MoveCursorToLineStart() {
  e.cursorX = 0
}

func (e *Editor) MoveCursorToLineEnd() {
  if len(e.lines[e.cursorY]) > 0 {
    e.cursorX = len(e.lines[e.cursorY]) - 1
  }
}

func (e *Editor) MoveCursorToNextWordStart() {
  i, j, found := FindNextWordStart(e.lines, e.cursorY, e.cursorX)
  if found {
    e.cursorY, e.cursorX = i, j
  }
}

func (e *Editor) MoveCursorToNextWordEnd() {
  i, j, found := FindNextWordEnd(e.lines, e.cursorY, e.cursorX)
  if found {
    e.cursorY, e.cursorX = i, j
  }
}

func (e *Editor) MoveCursorToPrevWordStart() {
  i, j, found := FindPrevWordStart(e.lines, e.cursorY, e.cursorX)
  if found {
    e.cursorY, e.cursorX = i, j
  }
}

func (e *Editor) HandleKeyboard(ch rune, key tcell.Key) bool {
  //fmt.Printf("- %V %V %V\n", ch, key, tcell.KeyBS)
  if e.mode == NORMAL {
    if e.buffCommand != "" && ch != rune(0) {
      e.HandleBufferedCommand(ch)
      e.UpdateText()
      return false
    } else {
      e.buffCommand = ""
    }

    if ch == rune(0) {
      switch key {
      case tcell.KeyDown:
        e.MoveCursorDown()
      case tcell.KeyUp:
        e.MoveCursorUp()
      case tcell.KeyRight:
        e.MoveCursorRight()
      case tcell.KeyLeft:
        e.MoveCursorLeft()
      case tcell.KeyCtrlR:
        state := e.history.Redo()
        if state != nil {
          e.lines, e.cursorX, e.cursorY = state.Unpack()
          e.fullText = ""
        }
      }
      e.UpdateText()
      return false
    }

    switch ch {
    case 'q':
      return true
    case 'j':
      e.MoveCursorDown()
    case 'k':
      e.MoveCursorUp()
    case 'l':
      e.MoveCursorRight()
    case 'h':
      e.MoveCursorLeft()
    case 'w':
      e.MoveCursorToNextWordStart()
    case 'e':
      e.MoveCursorToNextWordEnd()
    case 'b':
      e.MoveCursorToPrevWordStart()
    case 'v':
      e.selected = VisualSelect{e.cursorY, 1}
      e.tv.Highlight("visual")
      e.SetMode(VISUAL)
    case 'i':
      e.SaveHistory()
      e.SetMode(INSERT)
    case 'a':
      e.SaveHistory()
      e.MoveCursorRight()
      e.SetMode(INSERT)
    case 'x':
      e.SaveHistory()
      e.DeleteCharBefore(e.cursorY, e.cursorX + 1)
    case '0':
      e.MoveCursorToLineStart()
    case '$':
      e.MoveCursorToLineEnd()
    case 'o':
      e.SaveHistory()
      end := Max(0, len(e.lines[e.cursorY]))
      e.NewLineAfter(e.cursorY, end)
      e.MoveCursorDown()
      e.cursorX = 0
      e.SetMode(INSERT)
    case 'O':
      e.SaveHistory()
      e.NewLineAfter(e.cursorY, 0)
      e.cursorX = 0
      e.SetMode(INSERT)
    case 'p':
      e.SaveHistory()
      e.PasteYankedBuffer()
    case 'D':
      if e.cursorX < len(e.lines[e.cursorY]) {
        e.SaveHistory()
        e.yankedBuffer = e.lines[e.cursorY][e.cursorX:]
        e.lines[e.cursorY] = e.lines[e.cursorY][:e.cursorX]
        e.fullText = ""
      }
    case 'Y':
      if e.cursorX < len(e.lines[e.cursorY]) {
        e.yankedBuffer = e.lines[e.cursorY][e.cursorX:]
      }
    case 'r', 'd', 'y':
      e.buffCommand = string(ch)
    case 'u':
      if e.history.Current() == nil {
        e.SaveHistory()
        e.history.Undo()
      }

      state := e.history.Undo()
      if state != nil {
        e.lines, e.cursorX, e.cursorY = state.Unpack()
        e.fullText = ""
      }
    }
  } else if e.mode == VISUAL {
    switch ch {
    case 'q':
      e.tv.Highlight("cursor")
      e.SetMode(NORMAL)
    case 'j':
      if e.selected.start == e.cursorY {
        if e.selected.start + e.selected.size < len(e.lines) {
          e.selected.size += 1
        }
      } else if e.selected.start < e.cursorY {
        e.selected.start += 1
        e.selected.size  -= 1
      }
    case 'k':
      if e.selected.start == e.cursorY && e.selected.size > 1 {
        e.selected.size -= 1

      } else if e.selected.start > 0 {
        e.selected.start -= 1
        e.selected.size  += 1
      }
    }

    if key == tcell.KeyCtrlX {
      e.onExecute(e.GetSelectedText())
    }
  } else {
    switch key {
    case tcell.KeyESC:
      e.SetMode(NORMAL)
    case tcell.KeyDown:
      e.MoveCursorDown()
    case tcell.KeyUp:
      e.MoveCursorUp()
    case tcell.KeyRight:
      e.MoveCursorRight()
    case tcell.KeyLeft:
      e.MoveCursorLeft()
    case tcell.KeyBackspace2:
      linesLen, rowLen := len(e.lines), 0

      if e.cursorY > 0 {
        rowLen = len(e.lines[e.cursorY - 1])
      }

      e.DeleteCharBefore(e.cursorY, e.cursorX)

      if e.cursorY > 0 {
        if linesLen > len(e.lines) {
          e.MoveCursorUp()
          e.cursorX = rowLen
        } else {
          e.MoveCursorLeft()
        }
      } else {
        e.MoveCursorLeft()
      }
    case tcell.KeyCR:
      e.NewLineAfter(e.cursorY, e.cursorX)
      e.MoveCursorDown()
      e.cursorX = 0
    default:
      if key == 0 {
        e.InsertCharBefore(ch, e.cursorY, e.cursorX)
        e.MoveCursorRight()
      }
    }
  }

  e.UpdateText()
  return false
}

func (e *Editor) YankCurrentWord(fromStart bool) {
  if fromStart {
    yStart, xStart, ok1 := FindPrevWordStart(e.lines, e.cursorY, e.cursorX)
    yEnd,   xEnd,   ok2 := FindNextWordEnd(e.lines, e.cursorY, e.cursorX)

    if !ok1 || !ok2 || yStart != yEnd {
      // TODO: put some error
      return
    }

    if xEnd < len(e.lines) {
      e.yankedBuffer = e.lines[yStart][xStart: xEnd + 1]
    } else {
      e.yankedBuffer = e.lines[yStart][xStart:]
    }
  } else {
    yStart, xStart := e.cursorY, e.cursorX
    yEnd,   xEnd, ok := FindNextWordEnd(e.lines, e.cursorY, e.cursorX)

    if !ok || yStart != yEnd {
      // TODO: put some error
      return
    }

    if xEnd < len(e.lines) {
      e.yankedBuffer = e.lines[yStart][xStart: xEnd + 1]
    } else {
      e.yankedBuffer = e.lines[yStart][xStart:]
    }
  }
}

func (e *Editor) DeleteCurrentWord(fromStart bool) {
  if fromStart {
    yStart, xStart, ok1 := FindPrevWordStart(e.lines, e.cursorY, e.cursorX)
    yEnd,   xEnd,   ok2 := FindNextWordEnd(e.lines, e.cursorY, e.cursorX)

    if !ok1 || !ok2 || yStart != yEnd {
      // TODO: put some error
      return
    }

    if xEnd < len(e.lines[yStart]) {
      e.yankedBuffer = e.lines[yStart][xStart: xEnd + 1]
      e.lines[yStart] = e.lines[yStart][:xStart] + e.lines[yStart][xEnd + 1:]
      e.cursorX = xStart
    } else {
      e.yankedBuffer = e.lines[yStart][xStart:]
      e.lines[yStart] = e.lines[yStart][:xStart]
      e.cursorX = xStart
    }
  } else {
    yStart, xStart := e.cursorY, e.cursorX
    yEnd,   xEnd, ok := FindNextWordEnd(e.lines, e.cursorY, e.cursorX)

    if !ok || yStart != yEnd {
      // TODO: put some error
      return
    }

    if xEnd < len(e.lines[yStart]) {
      e.yankedBuffer = e.lines[yStart][xStart: xEnd + 1]
      e.lines[yStart] = e.lines[yStart][:xStart] + e.lines[yStart][xEnd + 1:]
      e.cursorX = xStart
    } else {
      e.yankedBuffer = e.lines[yStart][xStart:]
      e.lines[yStart] = e.lines[yStart][:xStart]
      e.cursorX = xStart
    }
  }
}

func (e *Editor) YankCurrentLine() {
  e.yankedBuffer = "\n" + e.lines[e.cursorY]
}

func (e *Editor) DelYankCurrentLine() {
  e.YankCurrentLine()
  if len(e.lines) > 1 {
    e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY + 1:]...)
  } else {
    e.lines = []string{""}
  }

  e.cursorY = Min(e.cursorY, len(e.lines) - 1)
}

func (e *Editor) PasteYankedBuffer() {
  if len(e.yankedBuffer) > 0 {
    if e.yankedBuffer[0] == '\n' {
      end := Max(0, len(e.lines[e.cursorY]))
      e.NewLineAfter(e.cursorY, end)
      e.lines[e.cursorY + 1] = e.yankedBuffer[1:]
    } else {
      e.InsertYankedWordAfter(e.cursorY, e.cursorX)
      e.cursorX += len(e.yankedBuffer)
    }
  }
}

func (e *Editor) HandleBufferedCommand(ch rune) {
  switch e.buffCommand[0] {
  case 'r':
    e.SaveHistory()
    e.DeleteCharBefore(e.cursorY, e.cursorX + 1)
    e.InsertCharBefore(ch, e.cursorY, e.cursorX)
    e.fullText = ""
  case 'y':
    switch ch {
    case 'i':
      if len(e.buffCommand) > 1 {
        e.buffCommand = ""
      } else {
        e.buffCommand = "yi"
      }
    case 'w':
      fromStart := len(e.buffCommand) > 1
      e.YankCurrentWord(fromStart)
    case '$':
      if e.cursorX < len(e.lines[e.cursorY]) {
        e.yankedBuffer = e.lines[e.cursorY][e.cursorX:]
      }
    case 'y':
      e.YankCurrentLine()
    }
  case 'd':
    switch ch {
    case 'i':
      if len(e.buffCommand) > 1 {
        e.buffCommand = ""
      } else {
        e.buffCommand = "di"
      }
      return
    case 'w':
      e.SaveHistory()
      fromStart := len(e.buffCommand) > 1
      e.DeleteCurrentWord(fromStart)
      e.fullText = ""
    case '$':
      if e.cursorX < len(e.lines[e.cursorY]) {
        e.SaveHistory()
        e.yankedBuffer = e.lines[e.cursorY][e.cursorX:]
        e.lines[e.cursorY] = e.lines[e.cursorY][:e.cursorX]
        e.fullText = ""
      }
    case 'd':
      e.SaveHistory()
      e.DelYankCurrentLine()
      e.fullText = ""
    }
  }

  e.buffCommand = ""
}

func (e *Editor) SetMode(m Mode) {
  e.mode = m
  e.onModeChanged(m)
}

func (e *Editor) SetModeChangeCb(cb func(Mode)) {
  e.onModeChanged = cb
}

func (e *Editor) SetExecuteCb(cb func(string)) {
  e.onExecute = cb
}

func (e *Editor) GetCursorLine() string {
  value := e.lines[e.cursorY]
  return value
}

func (e *Editor) GetFullText() string {
  var builder strings.Builder

  for _, line := range e.lines {
    builder.WriteString(line)
    builder.WriteString(" \n") // Adding space to be able to place cursor at the end of a line
  }
  return builder.String()
}

func Colorize(tt TokenType) bool {
  return tt == NUMBER || tt == STRING || tt == TYPE || tt == KEYWORD || tt == COMMENT
}

func InsertCursorTag(s string, at int) string {
  return s[:at] + "[\"cursor\"]" + string(s[at]) + "[\"\"]" + s[at + 1:] 
}

func (e *Editor) GenHighlight() {
  if e.fullText != "" {
    return
  }

  e.highlights = []Highlight{}

  text := e.GetFullText()
  e.fullText = text

  e.tokenizer.SetInput(text)

  normalStart := 0

  for !e.tokenizer.IsEnd() {
    token := e.tokenizer.NextToken()

    if Colorize(token.ttype) {
      if normalStart < token.start {
        hl := Highlight{ normalStart, token.start, White }
        e.highlights = append(e.highlights, hl)
      }

      if token.ttype == KEYWORD {
        hl := Highlight{ token.start, token.start + token.size, Violet }
        e.highlights = append(e.highlights, hl)
      } else if token.ttype == STRING {
        hl := Highlight{ token.start, token.start + token.size, Yellow }
        e.highlights = append(e.highlights, hl)
      } else if token.ttype == NUMBER {
        hl := Highlight{ token.start, token.start + token.size, Red }
        e.highlights = append(e.highlights, hl)
      } else if token.ttype == COMMENT {
        hl := Highlight{ token.start, token.start + token.size, Wheat }
        e.highlights = append(e.highlights, hl)
      } else if token.ttype == TYPE {
        hl := Highlight{ token.start, token.start + token.size, Turquoise }
        e.highlights = append(e.highlights, hl)
      }
      normalStart = token.start + token.size
    }
  }

  if normalStart < len(text) {
    hl := Highlight{ normalStart, len(text) - 1, White }
    e.highlights = append(e.highlights, hl)
  }
}

func (e *Editor) GetParsedText() string {
  e.GenHighlight()

  pos := 0
  for i := 0; i < e.cursorY; i++ {
    pos += len(e.lines[i]) + 2
  }
  pos += e.cursorX

  text := e.fullText

  parsedText := ""

  for _, hl := range e.highlights {
    value := text[hl.start: hl.end]

    if (e.mode != VISUAL) && pos >= hl.start && pos < hl.end {
      value = InsertCursorTag(value, pos - hl.start)
    }

    switch hl.color {
    case Yellow:
      value = "[yellow]" + value + "[white]"
    case Pink:
      value = "[hotpink]" + value + "[white]"
    case Red:
      value = "[tomato]" + value + "[white]"
    case Violet:
      value = "[violet]" + value + "[white]"
    case Blue:
      value = "[steelblue]" + value + "[white]"
    case Green:
      value = "[lawngreen]" + value + "[white]"
    case Orange:
      value = "[orangered]" + value + "[white]"
    case Gray:
      value = "[lightgray]" + value + "[white]"
    case Wheat:
      value = "[wheat]" + value + "[white]"
    case Turquoise:
      value = "[turquoise]" + value + "[white]"
    }

    parsedText += value
  }

  return parsedText
}

func (e *Editor) GetSelectedText() string {
  if e.mode == VISUAL {
    text := ""
    for i := e.selected.start; i < e.selected.start + e.selected.size; i++{
      text += e.lines[i] + " \n"
    }
    return text
  }
  return ""
}

func (e *Editor) InsertVisualTag(text string) string {
  lineCount := 1
  start, end := 0, 0
  startFound := false

  for i := 0; i < len(text); i++ {
    if text[i] == '\n' {
      lineCount++
    }

    if !startFound && e.selected.start == lineCount - 1 {
      startFound = true
      start = i
    }
    if e.selected.start + e.selected.size == lineCount - 1 {
      end = i
      break
    }
  }

  if start >= end {
    end = len(text) - 1
  }

  if len(text) == 0 || end >= len(text) || end < 0 || start < 0 || start >= len(text) || start >= end {
    panic(fmt.Sprintf("Err: len: %d, start: %d, end: %d, count: %d --- %v", len(text), start, end, lineCount, e.selected))
  }
  return text[:start] + `["visual"]` + text[start:end] + `[""]` + text[end:]
}

func (e *Editor) UpdateText() {
  text := e.GetParsedText()
  if e.mode == VISUAL {
    text = e.InsertVisualTag(text)
  }

  e.tv.SetText(text)
  e.tv.ScrollTo(e.cursorX, e.cursorY)
}
