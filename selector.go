package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "fmt"
  "strconv"
)

type Enumerable interface {
  GetItemName(index int) string
  Remove(index int) Enumerable
  Len() int
}

type Selector struct {
  tv *tview.TextView

  initialText string

  items Enumerable
  selected int
  cursor int

  onSelect func (int)
  onDelete func (int) bool
}

func NewSelector(tv *tview.TextView, items Enumerable, appendItems bool) *Selector {
  s := &Selector{
    tv: tv,
    onSelect: func(i int) {},
    onDelete: func(i int) bool { return true },
  }

  tv.SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
    if event.Rune() == 'j' {
      if s.cursor < s.items.Len() - 1 {
        s.cursor += 1
        s.tv.Highlight(strconv.Itoa(s.cursor))
      }
    } else if event.Rune() == 'k' {
      if s.cursor > 0 {
        s.cursor -= 1
        s.tv.Highlight(strconv.Itoa(s.cursor))
      }
    } else if event.Rune() == 'd' {
      if s.onDelete(s.cursor) {
        s.items = s.items.Remove(s.cursor)
        c := Max(0, Min(s.cursor, s.items.Len() - 1))

        s.SetItems(s.items)
        s.cursor = c
        s.tv.Highlight(strconv.Itoa(s.cursor))
      }
    } else if event.Key() == tcell.KeyCR {
      s.SelectItem(s.cursor)
      s.onSelect(s.selected)
    }

    return event
  })

  s.initialText = ""
  if appendItems {
    s.initialText = s.tv.GetText(false)
  }

  s.SetItems(items)
  return s
}

func Last(s string) byte {
  l := len(s)
  if l > 0 {
    return s[l - 1]
  }
  return 0
}

func (s *Selector) SetItems(items Enumerable) {
  s.selected = -1
  s.cursor = 0

  s.items = items

  text := ""
  i := 0
  for i < s.items.Len() {
    itemText := s.items.GetItemName(i)
    jumpLine := Last(itemText) == '\n'

    if jumpLine {
      itemText = itemText[:len(itemText)-1]
    }

    text += fmt.Sprintf(" [\"%d\"] %s [\"\"]\n", i, itemText)

    if jumpLine {
      text += "\n"
    }

    i++
  }

  s.tv.SetText(s.initialText + text)
  s.tv.Highlight("0")
}

func (s *Selector) SetSelectedCb(cb func(int)) {
  s.onSelect = cb
}

func (s *Selector) SetDeleteCb(cb func(int) bool) {
  s.onDelete = cb
}

func (s *Selector) SelectedItemIndex() int {
  return s.selected 
}

func (s *Selector) SelectItem(index int) {
  s.selected = index
         
  text := ""
  i := 0
  for i < s.items.Len() {
    itemText := s.items.GetItemName(i)
    jumpLine := Last(itemText) == '\n'

    if jumpLine {
      itemText = itemText[:len(itemText)-1]
    }

    if i != index {
      text += fmt.Sprintf(" [\"%d\"] %s [\"\"]\n", i, itemText)
    } else {
      text += fmt.Sprintf(" [\"%d\"][orangered] %s [white][\"\"]\n", i, itemText)
    }

    if jumpLine {
      text += "\n"
    }

    i++
  }

  s.tv.SetText(s.initialText + text)
}

