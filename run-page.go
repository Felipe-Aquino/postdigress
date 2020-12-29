package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "time"
)

type TableMode byte

const (
  NONE   TableMode = 0
  ROW    TableMode = 1
  COLUMN TableMode = 2
  CELL   TableMode = 3
)

type RunPage struct {
  editor *Editor

  table *tview.Table
  tableMode TableMode

  focusedType ComponentType

  focused   *tview.TextView
	modeName  *tview.TextView
	statusBar *tview.TextView
  layout    *tview.Grid
}

func NewRunPage(c *Context) *RunPage {
  rp := &RunPage{}

  rp.focusedType = MENU

  rp.editor = NewEditor()
  rp.editor.UpdateText()

  rp.tableMode = NONE

	rp.table = tview.NewTable().
		SetBorders(false).
		SetSeparator(tview.Borders.Vertical)

	rp.table.Select(0, 0).SetFixed(1, 1).
    SetDoneFunc(func (key tcell.Key) {
      if key == tcell.KeyEscape {
        c.Finish()
      }
    }).
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      if event.Rune() == 'm' {
        rp.tableMode = (rp.tableMode + 1) % 4

        rp.table.SetSelectable(
          rp.tableMode & 1 != 0,
          rp.tableMode & 2 != 0)

        rp.SetModeName()
      }
      return event
    })

  TableSetData(rp.table, []string{" "}, [][]string{}, false)

	rp.modeName = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	rp.focused = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

  rp.SetCompType(MENU)

  rp.editor.SetModeChangeCb(func (m Mode) {
    rp.SetModeName()
  })

  rp.editor.SetExecuteCb(func (query string) {
    if c.loading.waiting {
      return
    }

    c.loading.SetTextView(rp.statusBar)
    go c.loading.Init(c.app)

    go func () {
      startTime := time.Now()

      var queryResult QueryResult

      if len(query) > 0 {
        queryResult = GetQueryResult(c.db, query)
      } else {
        c.Enqueue(func () {
          rp.statusBar.SetText("Nothing to be done.")
        })
        return
      }

      duration := time.Now().Sub(startTime)

      c.loading.Close()

      if queryResult.err != nil {
        c.Enqueue(func () {
          rp.statusBar.SetText(queryResult.err.Error())
          TableSetData(rp.table, []string{}, [][]string{}, false)
        })
      } else if len(queryResult.columns) == 0 {
        c.Enqueue(func () {
          rp.statusBar.SetText("Finished in " + duration.String())
          TableSetData(rp.table, []string{}, [][]string{}, false)
        })
      } else {
        c.Enqueue(func () {
          rp.statusBar.SetText("Finished in " + duration.String())

          TableSetData(rp.table, queryResult.columns, queryResult.values, true)
        })
      }
    }()
  })

	rp.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

  rp.statusBar.SetText(" Nothing new.")

	rp.layout = tview.NewGrid().
		SetBorders(true).
		SetRows(1, -1, -2, 1).
		SetColumns(8, 8, -1).
		AddItem(c.menuBar,    0, 0, 1, 3, 0, 0, true).
		AddItem(rp.editor.tv, 1, 0, 1, 3, 0, 0, false).
		AddItem(rp.table,     2, 0, 1, 3, 0, 0, false).
		AddItem(rp.focused,   3, 0, 1, 1, 0, 0, false).
		AddItem(rp.modeName,  3, 1, 1, 1, 0, 0, false).
		AddItem(rp.statusBar, 3, 2, 1, 1, 0, 0, false)

  rp.layout.
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      if rp.focusedType == EDITOR && rp.editor.mode == NORMAL {
        if event.Rune() == 'q' {
          rp.SetCompType(MENU)
          c.FocusMenu()
        } else if event.Key() == tcell.KeyCtrlT {
          rp.SetCompType(TABLE)
          c.SetFocus(rp.table)
        }
      } else if rp.focusedType == TABLE {
        if event.Rune() == 'q' {
          rp.SetCompType(MENU)
          c.FocusMenu()
        } else if event.Key() == tcell.KeyCtrlE {
          rp.SetCompType(EDITOR)
          c.SetFocus(rp.editor.tv)
        }
      } else if rp.focusedType == MENU {
        c.HandleMenuKeyInput(event)
      }

      return event
    })

  return rp
}

func (rp *RunPage) SetCompType(t ComponentType) {
  rp.focusedType = t

  switch t {
  case MENU:
    rp.editor.tv.Highlight("")
    rp.focused.SetText(" MENU ")
  case EDITOR:
    rp.focused.SetText(" EDITOR ")
    rp.editor.tv.Highlight("cursor")
  case TABLE:
    rp.editor.tv.Highlight("")
    rp.focused.SetText(" TABLE ")
  default:
    rp.editor.tv.Highlight("")
    rp.focused.SetText(" ??? ")
  }

  rp.SetModeName()
}

func (rp *RunPage) SetModeName() {
  if rp.focusedType == TABLE {
    switch rp.tableMode {
    case NONE:
      rp.modeName.SetText(" NONE ")
    case ROW:
      rp.modeName.SetText(" ROW ")
    case COLUMN:
      rp.modeName.SetText(" COLUMN ")
    case CELL:
      rp.modeName.SetText(" CELL ")
    }

  } else if rp.focusedType == EDITOR {
    switch rp.editor.mode {
    case NORMAL:
      rp.modeName.SetText(" NORMAL ")
    case INSERT:
      rp.modeName.SetText(" INSERT ")
    case VISUAL:
      rp.modeName.SetText(" VISUAL ")
    }
  } else {
    rp.modeName.SetText(" MENU ")
  }
}

func (rp *RunPage) SetStatus(msg string) {
  rp.statusBar.SetText(msg)
}

func (rp *RunPage) Layout() tview.Primitive {
  return rp.layout
}

