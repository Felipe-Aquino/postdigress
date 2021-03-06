package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
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

  text Text
  mode Mode

  useLineNumbers bool // enable line numbers
  numbersShift int // number of charaters shifted to give space to line numbers

  cursorX, cursorY int

  tokenizer *Tokenizer
  highlights []Highlight
  fullText []rune
  modified bool

  onModeChanged func(Mode)
  onExecute func(string)

  selected VisualSelect

  buffCommand string // buffer that save a multi-letter command
  yankedLines Text  // used to save copied/deleted text
}

func NewEditor() *Editor {
  e := &Editor{
    tv: tview.NewTextView(),
    text: WrapLines(
      "-- SQL ", 
      "",
    ),
    mode: NORMAL,
    onModeChanged: func(m Mode) {
    },
    onExecute: func(s string) {
    },
  }

  e.history = NewDumbHistory(5)

  e.tokenizer = NewTokenizer()
  e.modified = true

	e.tv.
    SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true).
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      if event.Key() == tcell.KeyRune {
        e.HandleKeyboard(event.Rune(), 0)
      } else {
        e.HandleKeyboard(rune(0), event.Key())
      }
      return nil
    })

  return e
}

func (e *Editor) EnableLineNumber(enable bool) {
  e.useLineNumbers = enable
  e.modified = true
  e.UpdateText()
}

func (e *Editor) SaveHistory() {
  es := NewEditorState(e.text.Clone(), e.cursorX, e.cursorY)
  e.history.Add(es)
}

func (e *Editor) NewLineAt(row int, save bool) {
  if save {
    e.SaveHistory()
  }

  e.text = e.text.InsertLines(row, "")
  e.modified = true
}

func (e *Editor) InsertYankedAfter(row, col int) {
  e.SaveHistory()
  col = Min(col + 1, e.text.LineLen(row))
  e.text = e.text.InsertAt(row, col, e.yankedLines)
  e.modified = true
}

func (e *Editor) InsertCharBefore(ch rune, row, col int) {
  e.text = e.text.InsertAt(row, col, WrapLines(string(ch)))
  e.modified = true
}

func (e *Editor) DeleteCharBefore(row, col int) {
  if col == 0 {
    if row > 0 {
      lineLen := e.text.LineLen(row - 1)
      e.text = e.text.DeleteRange(row - 1, lineLen, row, -1)
    }
  } else {
    e.text = e.text.DeleteRange(row, col - 1, row, col - 1)
  }
  e.modified = true
}

func (e *Editor) MoveCursorUp() {
  if e.cursorY > 0 {
    e.cursorY--

    lineLen := e.text.LineLen(e.cursorY)
    if e.cursorX > lineLen - 1 {
      e.cursorX = Max(0, lineLen - 1)
    }
  }
}

func (e *Editor) MoveCursorDown() {
  if e.cursorY < e.text.Len() - 1 {
    e.cursorY++

    lineLen := e.text.LineLen(e.cursorY)
    if e.cursorX > lineLen - 1 {
      e.cursorX = Max(0, lineLen - 1)
    }
  }
}

func (e *Editor) MoveCursorRight() {
  if e.cursorX < e.text.LineLen(e.cursorY) {
    e.cursorX++
  }
}

func (e *Editor) MoveCursorLeft() {
  if e.cursorX > 0 {
    e.cursorX--
  }
}

func (e *Editor) MoveCursorToNextWordStart() {
  i, j, found := FindNextWordStart(e.text, e.cursorY, e.cursorX)
  if found {
    e.cursorY, e.cursorX = i, j
  }
}

func (e *Editor) MoveCursorToNextWordEnd() {
  i, j, found := FindNextWordEnd(e.text, e.cursorY, e.cursorX)
  if found {
    e.cursorY, e.cursorX = i, j
  }
}

func (e *Editor) MoveCursorToPrevWordStart() {
  i, j, found := FindPrevWordStart(e.text, e.cursorY, e.cursorX)
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
          e.text, e.cursorX, e.cursorY = state.Unpack()
          e.modified = true
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
      // TODO Verify, maybe the text is not modified, but history is saved
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
      e.cursorX = 0
    case '$':
      if e.text.LineLen(e.cursorY) != 0 {
        e.cursorX = e.text.LineLen(e.cursorY) - 1
      }
    case 'o':
      e.NewLineAt(e.cursorY + 1, true)
      e.MoveCursorDown()
      e.cursorX = 0
      e.SetMode(INSERT)
    case 'O':
      e.NewLineAt(e.cursorY, true)
      e.cursorX = 0
      e.SetMode(INSERT)
    case 'p':
      e.PasteYankedBuffer()
    case 'D':
      lineLen := e.text.LineLen(e.cursorY)
      if e.cursorX < lineLen {
        e.SaveHistory()
        e.yankedLines = e.text.DeleteSubStrAt(e.cursorY, e.cursorX, lineLen)
        e.modified = true
      }
    case 'Y':
      lineLen := e.text.LineLen(e.cursorY)
      if e.cursorX < lineLen {
        e.yankedLines = e.text.SubStrAt(e.cursorY, e.cursorX, lineLen)
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
        e.text, e.cursorX, e.cursorY = state.Unpack()
        e.modified = true
      }
    }
  } else if e.mode == VISUAL {
    switch ch {
    case 'q':
      e.tv.Highlight("cursor")
      e.SetMode(NORMAL)
    case 'j':
      if e.selected.start == e.cursorY {
        if e.selected.start + e.selected.size < e.text.Len() {
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
      nLines, rowLen := e.text.Len(), 0

      if e.cursorY > 0 {
        rowLen = e.text.LineLen(e.cursorY - 1)
      }

      e.DeleteCharBefore(e.cursorY, e.cursorX)

      if e.cursorY > 0 {
        if nLines > e.text.Len() {
          e.MoveCursorUp()
          e.cursorX = rowLen
        } else {
          e.MoveCursorLeft()
        }
      } else {
        e.MoveCursorLeft()
      }
    case tcell.KeyCR:
      e.NewLineAt(e.cursorY + 1, false)
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
    yStart, xStart, ok1 := FindPrevWordStart(e.text, e.cursorY, e.cursorX)
    yEnd,   xEnd,   ok2 := FindNextWordEnd(e.text, e.cursorY, e.cursorX)

    if !ok1 || !ok2 || yStart != yEnd {
      // TODO: put some error
      return
    }

    xEnd = Min(xEnd + 1, e.text.LineLen(yEnd))
    e.yankedLines = e.text.SubStrAt(yStart, xStart, xEnd)

  } else {
    yStart, xStart := e.cursorY, e.cursorX
    yEnd,   xEnd, ok := FindNextWordEnd(e.text, e.cursorY, e.cursorX)

    if !ok || yStart != yEnd {
      // TODO: put some error
      return
    }

    xEnd = Min(xEnd + 1, e.text.LineLen(yEnd))
    e.yankedLines = e.text.SubStrAt(yStart, xStart, xEnd)
  }
}

func (e *Editor) DeleteCurrentWord(fromStart bool) {
  e.SaveHistory()

  if fromStart {
    yStart, xStart, ok1 := FindPrevWordStart(e.text, e.cursorY, e.cursorX)
    yEnd,   xEnd,   ok2 := FindNextWordEnd(e.text, e.cursorY, e.cursorX)

    if !ok1 || !ok2 || yStart != yEnd {
      // TODO: put some error
      return
    }

    xEnd = Min(xEnd + 1, e.text.LineLen(yEnd))
    e.yankedLines = e.text.DeleteSubStrAt(yStart, xStart, xEnd)
    e.cursorX = xStart

  } else {
    yStart, xStart := e.cursorY, e.cursorX
    yEnd,   xEnd, ok := FindNextWordEnd(e.text, e.cursorY, e.cursorX)

    if !ok || yStart != yEnd {
      // TODO: put some error
      return
    }

    xEnd = Min(xEnd + 1, e.text.LineLen(yEnd))
    e.yankedLines = e.text.DeleteSubStrAt(yStart, xStart, xEnd)
    e.cursorX = xStart
  }

  e.modified = true
}

func (e *Editor) DeleteOrYankInside(ch rune, del bool) {
  beginCh, endCh := ch, ch
  switch ch {
  case '(', ')':
    beginCh, endCh = '(', ')'
  case '[', ']':
    beginCh, endCh = '[', ']'
  case '{', '}':
    beginCh, endCh = '{', '}'
  }

  x, y := e.cursorX, e.cursorY
  multiline := ch != '\'' && ch != '"'

  yStart, xStart, found := FindCharBackwards(e.text, beginCh, y, x, multiline)

  if !found {
    return
  }

  yEnd,   xEnd,   found := FindCharForwards(e.text,  endCh,   y, x, multiline)

  if !found {
    return
  }

  xStart += 1
  xEnd -= 1

  if yStart == yEnd && xStart > xEnd {
    return
  }

  e.yankedLines = e.text.CopyRange(yStart, xStart, yEnd, xEnd)

  if del {
    e.SaveHistory()
    e.text = e.text.DeleteRange(yStart, xStart, yEnd, xEnd)
    e.modified = true
    e.cursorX = xStart
  }
}

func (e *Editor) SetYanked(txt Text) {
  e.yankedLines = txt
}

func (e *Editor) YankCurrentLine() {
  e.yankedLines = WrapLinesR(Line{}, e.text.Line(e.cursorY).Clone())
}

func (e *Editor) DelYankCurrentLine() {
  e.YankCurrentLine()
  e.SaveHistory()

  if e.text.Len() > 1 {
    e.text = e.text.DeleteLines(e.cursorY, e.cursorY)
  } else {
    e.text = WrapLinesR(Line{})
  }

  e.cursorY = Min(e.cursorY, e.text.Len()- 1)
  e.modified = true
}

func (e *Editor) PasteYankedBuffer() {
  if e.yankedLines.Len() > 0 {
    if e.yankedLines.LineLen(0) == 0 {
      end := e.text.LineLen(e.cursorY)
      e.InsertYankedAfter(e.cursorY, end)
      e.cursorY += 1
    } else {
      e.InsertYankedAfter(e.cursorY, e.cursorX)
      e.cursorX += e.yankedLines.LineLen(0)
    }
  }
}

func (e *Editor) HandleBufferedCommand(ch rune) {
  switch e.buffCommand[0] {
  case 'r':
    e.SaveHistory()
    e.text.ReplaceChar(e.cursorY, e.cursorX, ch)
    e.modified = true
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
      lineLen := e.text.LineLen(e.cursorY)
      if e.cursorX < lineLen {
        e.yankedLines = e.text.SubStrAt(e.cursorY, e.cursorX, lineLen)
      }
    case 'y':
      e.YankCurrentLine()
    case '\'', '"', '(', ')', '[', ']', '{', '}':
      if len(e.buffCommand) > 1 {
        e.DeleteOrYankInside(ch, false)
      }
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
      fromStart := len(e.buffCommand) > 1
      e.DeleteCurrentWord(fromStart)
    case '$':
      lineLen := e.text.LineLen(e.cursorY)
      if e.cursorX < lineLen {
        e.SaveHistory()
        e.yankedLines = e.text.DeleteSubStrAt(e.cursorY, e.cursorX, lineLen)
        e.modified = true
      }
    case 'd':
      e.DelYankCurrentLine()
    case '\'', '"', '(', ')', '[', ']', '{', '}':
      if len(e.buffCommand) > 1 {
        e.DeleteOrYankInside(ch, true)
      }
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

func (e *Editor) SetText(text Text) {
  e.SaveHistory()

  e.cursorX, e.cursorY = 0, 0

  if len(text) == 0 {
    text = WrapLines("")
  }

  e.text = text
  e.modified = true
  e.UpdateText()
}

func NumDig(n int) int {
  count := 0

  for n != 0 {
    n = n / 10
    count += 1
  }
  return count
}

func (e *Editor) GetFullText() []rune {
  result := []rune{}

  strPad := ""

  if e.useLineNumbers {
    e.numbersShift = Max(2, NumDig(e.text.Len())) + 2
    strPad = fmt.Sprintf("%%%dv ", e.numbersShift - 1)
  } else {
    e.numbersShift = 0
  }

  for i, line := range e.text {
    if e.useLineNumbers {
      result = append(result, []rune(fmt.Sprintf(strPad, i + 1))...)
    }

    result = append(result, line...)
    result = append(result, ' ', '\n') // Adding space to be able to place cursor at the end of a line
  }

  return result
}

func Colorize(tt TokenType) bool {
  return tt == NUMBER || tt == STRING || tt == TYPE || tt == KEYWORD || tt == COMMENT
}

func (e *Editor) GenHighlight() {
  if !e.modified {
    return
  }
  e.modified = false

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

func Tint(value []rune, color string) []rune {
  return append([]rune("[" + color + "]"), append(value, []rune("[white]")...)...)
}

func (e *Editor) GetParsedText() []rune {
  e.GenHighlight()

  pos := 0
  for i := 0; i < e.cursorY; i++ {
    pos += e.text.LineLen(i) + e.numbersShift + 2
  }
  pos += e.cursorX + e.numbersShift

  text := e.fullText

  parsedText := []rune{}

  for _, hl := range e.highlights {
    value := append([]rune{}, text[hl.start: hl.end]...)

    if (e.mode != VISUAL) && pos >= hl.start && pos < hl.end {
      value = InsertCursorTag(value, pos - hl.start)
    }

    switch hl.color {
    case Yellow:
      value = Tint(value, "yellow")
    case Pink:
      value = Tint(value, "hotpink")
    case Red:
      value = Tint(value, "tomato")
    case Violet:
      value = Tint(value, "violet")
    case Blue:
      value = Tint(value, "steelblue")
    case Green:
      value = Tint(value, "lawngreen")
    case Orange:
      value = Tint(value, "orangered")
    case Gray:
      value = Tint(value, "lightgray")
    case Wheat:
      value = Tint(value, "wheat")
    case Turquoise:
      value = Tint(value, "turquoise")
    }

    parsedText = append(parsedText, value...)
  }

  return parsedText
}

func (e *Editor) GetSelectedText() string {
  if e.mode == VISUAL {
    text := ""
    for i := e.selected.start; i < e.selected.start + e.selected.size; i++ {
      text += e.text.Line(i).String() + " \n"
    }
    return text
  }
  return ""
}

func (e *Editor) InsertVisualTag(text []rune) []rune {
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

  result := append([]rune{}, text[:start]...)
  result = append(result, []rune(`["visual"]`)...)
  result = append(result, text[start:end]...)
  result = append(result, []rune(`[""]`)...)
  result = append(result, text[end:]...)
  return result
}

func (e *Editor) UpdateText() {
  text := e.GetParsedText()
  if e.mode == VISUAL {
    text = e.InsertVisualTag(text)
  }

  e.tv.SetText(string(text))
  e.tv.ScrollToHighlight()
}
