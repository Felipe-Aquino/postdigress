package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "fmt"
  "hash/fnv"
)

func IsAlpha(r rune) bool {
  return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func IsAlnum(r rune) bool {
  return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func IsDigit(r rune) bool {
  return r >= '0' && r <= '9'
}

func IsSpace(r rune) bool {
  return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func IsSpecialChar(r rune) bool {
  return !IsAlnum(r) && !IsSpace(r) && r != '_'
}

func IsWordChar(r rune) bool {
  return IsAlnum(r) || r == '_'
}

func Tern(cond bool, v1, v2 int) int {
  if cond {
    return v1
  }
  return v2
}

func Hash(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}

type Rect struct {
  x, y, width, height int
}

func (r Rect) Unpack() (int, int, int, int) {
  return r.x, r.y, r.width, r.height
}

func (r Rect) YMax() int {
  return r.y + r.height
}

func (r Rect) XMax() int {
  return r.x + r.width
}

func Min(a, b int) int {
  if a > b {
    return b
  }
  return a
}

func Max(a, b int) int {
  if a < b {
    return b
  }
  return a
}

func DigitNumber(n int) int {
  if n < 10 {
    return 1
  } else if n < 100 {
    return 2
  }
  return 3
}

func StrClip(s string, max int) string {
  if len(s) > max {
    return s[:max]
  }
  return s
}

func StrClamp(s string, min, max int) string {
  size := len(s)
  if size > min {
    if size > max {
      return s[min:max]
    }
    return s[min:]
  }
  return ""
}

func RuneSliceClamp(s []rune, min, max int) []rune {
  size := len(s)
  if size > min {
    if size > max {
      return s[min:max]
    }
    return s[min:]
  }
  return []rune{}
}

// Split string in N pieces
func StrSplitInN(s string, n int) ([]string, bool) {
	pieces := []string{}

  if n < 1 {
    return pieces, true
  }

	nPieces := len(s) / n

	start := 0
	for i := 0; i < nPieces; i++ {
		pieces = append(pieces, s[start:(i + 1)*n])
		start = (i + 1) * n
	}

	if start < len(s) {
		pieces = append(pieces, s[start:])
	}

  return pieces, false
}

func U8Charsize(ch int) int {
  if ch <= 0x7F {
    return 1
  } else if ch <= 0xE0 {
    return 2
  } else if ch <= 0xF0 {
    return 3
  }
  return 4
}

func SSMap(vs [][]string, f func([]string) string) []string {
  vsm := make([]string, len(vs))
  for i, v := range vs {
    vsm[i] = f(v)
  }
  return vsm
}

///* tview *///

func TableSetData(table *tview.Table, fields []string, values [][]string, showNumbers bool) {
	cols, rows := len(fields), len(values)

  table.Clear()

  if showNumbers {
    table.SetCell(0, 0,
      tview.NewTableCell(" # ").
        SetTextColor(tcell.ColorYellow).
        SetAlign(tview.AlignCenter))
  }

  for j := 0; j < cols; j++ {
    table.SetCell(0, Tern(showNumbers, j + 1, j),
      tview.NewTableCell(" " + fields[j] + " ").
        SetTextColor(tcell.ColorYellow).
        SetAlign(tview.AlignCenter))
  }

	for r := 0; r < rows; r++ {
    if showNumbers {
      table.SetCell(r + 1, 0,
        tview.NewTableCell(fmt.Sprintf("%d", r)).
          SetTextColor(tcell.ColorYellow).
          SetAlign(tview.AlignCenter))
    }

    cols = Min(cols, len(values[r]))
		for c := 0; c < cols; c++ {
			table.SetCell(r + 1, Tern(showNumbers, c + 1, c),
				tview.NewTableCell(" " + values[r][c] + " ").
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
		}
	}

  if rows == 0 {
    c := Max(0, (cols - 1) / 2)
    table.SetCell(2, c,
      tview.NewTableCell(" No Data ").
        SetTextColor(tcell.ColorBlue).
        SetAlign(tview.AlignCenter))

    table.SetCell(4, c,
      tview.NewTableCell(" ").
        SetTextColor(tcell.ColorBlue).
        SetAlign(tview.AlignCenter))
  }
}

func GetFormInputValue(form *tview.Form, index int) string {
  field, ok := form.GetFormItem(index).(*tview.InputField)
  if ok {
    return field.GetText()
  }
  return ""
}

func GetFormCheckValue(form *tview.Form, index int) bool {
  check, ok := form.GetFormItem(index).(*tview.Checkbox)
  if ok {
    return check.IsChecked()
  }
  return false
}

func SetFormInputValue(form *tview.Form, index int, value string) {
  field, ok := form.GetFormItem(index).(*tview.InputField)
  if ok {
    field.SetText(value)
  }
}

func SetFormCheckValue(form *tview.Form, index int, value bool) {
  check, ok := form.GetFormItem(index).(*tview.Checkbox)
  if ok {
    check.SetChecked(value)
  }
}

func GetFormKeyHandler(
  app *tview.Application,
  form *tview.Form, items,
  buttons int) func (event *tcell.EventKey) *tcell.EventKey {

  return func (event *tcell.EventKey) *tcell.EventKey {
    idx, btn := form.GetFocusedItemIndex()

    if btn == -1 {
      switch event.Key() {
      case tcell.KeyDown, tcell.KeyCtrlJ, tcell.KeyTab, tcell.KeyCtrlN:
        idx = idx + 1
        if idx < items {
          item := form.GetFormItem(idx)
          app.SetFocus(item)
        } else {
          item := form.GetButton(0)
          app.SetFocus(item)
        }

        return nil
      case tcell.KeyUp, tcell.KeyCtrlK, tcell.KeyCtrlP:
        idx--
        if idx < 0 {
          item := form.GetButton(0)
          app.SetFocus(item)
          return nil
        }
        item := form.GetFormItem(idx)
        app.SetFocus(item)
      }
    } else {
      switch event.Key() {
      case tcell.KeyDown, tcell.KeyCtrlJ, tcell.KeyCtrlN:
        item := form.GetFormItem(0)
        app.SetFocus(item)
      case tcell.KeyUp, tcell.KeyCtrlK, tcell.KeyCtrlP:
        item := form.GetFormItem(items - 1)
        app.SetFocus(item)
      case tcell.KeyLeft, tcell.KeyCtrlH, tcell.KeyRight, tcell.KeyCtrlL:
        item := form.GetButton((btn + 1) % buttons)
        app.SetFocus(item)
      case tcell.KeyTab:
        if btn == 0 {
          item := form.GetButton((btn + 1) % buttons)
          app.SetFocus(item)
        } else {
          item := form.GetFormItem(0)
          app.SetFocus(item)
        }
        return nil
      }
    }
    return event
  }
}

