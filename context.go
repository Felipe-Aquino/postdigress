package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"

	"database/sql"

	_ "github.com/lib/pq"
)

type ComponentType byte

const (
  MENU ComponentType = iota
  EDITOR
  TABLE

  DATABASE
  COLUMNS
  INDEXES
)

type SelectedMenu byte
const (
  RUN_MENU     SelectedMenu = iota
  STRUCT_MENU
)

type Context struct {
  db *sql.DB

  app *tview.Application
  info *DBInfo

	mainPages *tview.Pages
  menuBar   *tview.TextView

  selectedMenu SelectedMenu

  runPage    *RunPage
  structPage *StructPage

  loading *Loading
}

func (c *Context) Finish() {
  if c.loading != nil && c.loading.waiting {
    c.loading.Close()
  }

  if c.app != nil {
    c.app.Stop()
  }
}

func (c *Context) FocusMenu() {
  if c.menuBar != nil {
    c.app.SetFocus(c.menuBar)
  }
}

func (c *Context) SetFocus(p tview.Primitive) {
  c.app.SetFocus(p)
}

func (c *Context) HandleMenuKeyInput(event *tcell.EventKey) {
  if event.Rune() == 'q' {
    c.Finish()

  } else if event.Rune() == 'e' && c.selectedMenu != RUN_MENU {
    c.menuBar.Highlight("0")
    c.selectedMenu = RUN_MENU

  } else if event.Rune() == 's' && c.selectedMenu != STRUCT_MENU {
    c.menuBar.Highlight("1")
    c.selectedMenu = STRUCT_MENU

  } else if c.selectedMenu == RUN_MENU {
    if event.Key() == tcell.KeyCtrlE || event.Rune() == '0' {
      c.runPage.SetCompType(EDITOR)
      c.SetFocus(c.runPage.editor.tv)

    } else if event.Key() == tcell.KeyCtrlT || event.Rune() == '9' {
      c.runPage.SetCompType(TABLE)
      c.SetFocus(c.runPage.table)
    }

  } else if c.selectedMenu == STRUCT_MENU {
    if event.Rune() == 'd' {
      c.structPage.SetCompType(DATABASE)
      c.SetFocus(c.structPage.dbSelect)

    } else if event.Rune() == 'c' {
      c.structPage.SetCompType(COLUMNS)
      c.SetFocus(c.structPage.columnsTable)

    } else if event.Rune() == 'i' {
      c.structPage.SetCompType(INDEXES)
      c.SetFocus(c.structPage.indexesTable)
    }
  }
}

func (c *Context) Enqueue(fn func ()) {
  c.app.QueueUpdateDraw(fn)
}
