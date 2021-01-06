package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "time"
  "fmt"

  "strconv"
  "strings"
	"database/sql"
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

  focused  *tview.TextView
	modeName *tview.TextView
  layout   *tview.Grid

	status  *Status
  command *Command
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

    c.loading.SetTextView(rp.status.tv)
    go c.loading.Init(c.app)

    go func () {
      startTime := time.Now()

      var queryResult QueryResult

      if len(query) > 0 {
        queryResult = GetQueryResult(c.db, query)
      } else {
        c.Enqueue(func () {
          rp.status.SetText("Nothing to be done.")
        })
        return
      }

      duration := time.Now().Sub(startTime)

      c.loading.Close()

      if queryResult.err != nil {
        c.Enqueue(func () {
          rp.status.SetText(queryResult.err.Error())
          TableSetData(rp.table, []string{}, [][]string{}, false)
        })
      } else if len(queryResult.columns) == 0 {
        c.Enqueue(func () {
          rp.status.SetText("Finished in " + duration.String())
          TableSetData(rp.table, []string{}, [][]string{}, false)
        })
      } else {
        c.Enqueue(func () {
          rp.status.SetText("Finished in " + duration.String())

          TableSetData(rp.table, queryResult.columns, queryResult.values, true)
        })
      }
    }()
  })

  rp.command = NewCommand()
  rp.command.Register("yank", rp.Yank)
  rp.command.Register("yank-line", rp.YankLine)
  rp.command.Register("import", rp.Import)
  rp.command.Register("export", rp.Export)
  rp.command.Register("enable", rp.Enable)

  rp.command.Register("table-get", rp.TableGet)
  rp.command.Register("select-for", rp.YankSelectFor)
  rp.command.Register("insert-for",
    func(t string) string { return YankInsertFor(rp, c.db, t) }) 

  rp.status = NewStatus()
  rp.status.ChangeStartString(":")
  rp.status.SetMode(Show)

  rp.status.SetText("Nothing new.")

  rp.status.SetEnterCb(func(s string) {
    returned, err := rp.command.Run(s)

    rp.status.SetMode(Show)

    if err != nil {
      rp.status.SetText(err.Error())
    } else {
      rp.status.SetText(returned)
    }
    // Change Focus

    rp.SetCompType(EDITOR)
    c.SetFocus(rp.editor.tv)
  })

  rp.status.SetCancelCb(func() {
    // Change Focus
    rp.SetCompType(EDITOR)
    c.SetFocus(rp.editor.tv)
  })

	rp.layout = tview.NewGrid().
		SetBorders(true).
		SetRows(1, -2, -3, 1).
		SetColumns(8, 8, -1).
		AddItem(c.menuBar,    0, 0, 1, 3, 0, 0, true).
		AddItem(rp.editor.tv, 1, 0, 1, 3, 0, 0, false).
		AddItem(rp.table,     2, 0, 1, 3, 0, 0, false).
		AddItem(rp.focused,   3, 0, 1, 1, 0, 0, false).
		AddItem(rp.modeName,  3, 1, 1, 1, 0, 0, false).
		AddItem(rp.status.tv, 3, 2, 1, 1, 0, 0, false)

  rp.layout.
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      if rp.focusedType == EDITOR && rp.editor.mode == NORMAL {
        if event.Rune() == 'q' {
          rp.SetCompType(MENU)
          c.FocusMenu()
        } else if event.Rune() == ':' {
          rp.status.SetMode(Prompt)
          rp.SetCompType(COMMAND)
          c.SetFocus(rp.status.tv)
          return nil
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
  case COMMAND:
    rp.editor.tv.Highlight("")
    rp.focused.SetText(" PROMPT ")
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
  } else if rp.focusedType == MENU {
    rp.modeName.SetText(" MENU ")
  } else {
    rp.modeName.SetText("  ")
  }
}

func (rp *RunPage) SetStatus(msg string) {
  rp.status.SetText(msg)
}

func (rp *RunPage) Layout() tview.Primitive {
  return rp.layout
}

func (rp *RunPage) Enable(item string, enable bool) string {
  switch item {
  case "numbers":
    rp.editor.EnableLineNumber(enable)
  default:
    return item + " is undefined."
  }

  if enable {
    return item + " enabled."
  }

  return item + " disabled."
}

func (rp *RunPage) Yank(s string) string {
  rp.editor.SetYanked(WrapLines(s))
  return s
}

func (rp *RunPage) YankLine(s string) string {
  rp.editor.SetYanked(WrapLines("", s))
  return s
}

func (rp *RunPage) Import(path string) string {
  file, err := ReadFile(path)

  if err != nil {
    return err.Error()
  }

  rp.editor.SetText(TextFromString(string(file)))

  return path + " imported."
}

func (rp *RunPage) Export(path string) string {
  data := rp.editor.text.String()
  err := WriteFile(path, data)

  if err != nil {
    return err.Error()
  }

  return "Exported to " + path
}

func (rp *RunPage) TableGet(row, col string) string {
  nRow, err1 := strconv.Atoi(row)
  nCol, err2 := strconv.Atoi(col)

  if err1 != nil && err2 != nil {
    return "Could not get the value(s)"
  }
  if row == "-" && err2 == nil {
    nRow = rp.table.GetRowCount()

    result := "("
    for i := 0; i < nRow; i++ {
      result += "'" + rp.table.GetCell(i, nCol).Text + "'"

      if i != nRow - 1 {
        result += ", "
      }
    }
    result += ")"

    rp.editor.SetYanked(WrapLines(result))
    return "Values yanked."
  }
  if col == "-" && err1 == nil {
    nCol = rp.table.GetColumnCount()

    result := "("
    for i := 0; i < nCol; i++ {
      result += "'" + rp.table.GetCell(nRow, i).Text + "'"

      if i != nCol - 1 {
        result += ", "
      }
    }
    result += ")"

    rp.editor.SetYanked(WrapLines(result))
    return "Values yanked."
  }

  result := rp.table.GetCell(nRow, nCol).Text
  rp.editor.SetYanked(WrapLines(result))
  return "Value yanked."
}

func (rp *RunPage) YankSelectFor(table string) string {
  text := WrapLines("", "SELECT * FROM " + table + ";")
  rp.editor.SetYanked(text)

  return "Select text yanked."
}

func YankInsertFor(rp *RunPage, db *sql.DB, table string) string {
  query := `select column_name, data_type
            from information_schema.columns
            where table_name = '` + table + "';"

  result := GetQueryResult(db, query)

  columnNames := "("
  valuePrototye := "("
  total := len(result.values)

  if total == 0 {
    return "Couldn't generate query for " + table
  }

  for i, e := range result.values {
    columnType := e[1]

    columnNames += e[0]

    if strings.HasPrefix(columnType, "bool") {
      valuePrototye += "false"

    } else if strings.Contains(columnType, "char") {
      valuePrototye += "''"

    } else if strings.Contains(columnType, "time") ||
              strings.Contains(columnType, "date") {
      valuePrototye += "'YYYY-MM-DD HH:mm:ssZ'"

    } else if strings.Contains(columnType, "int") {
      valuePrototye += "0"

    } else if strings.HasPrefix(columnType, "double")  ||
              strings.HasPrefix(columnType, "numeric") ||
              strings.Contains(columnType, "real")     ||
              strings.Contains(columnType, "decimal") {
      valuePrototye += "0.0"

    }

    if i != total - 1 {
      columnNames   += ", "
      valuePrototye += ", "
    }
  }

  valuePrototye += ")"
  columnNames   += ")"

  query = fmt.Sprintf("INSERT INTO %s %s VALUES", table, columnNames)

  rp.editor.SetYanked(WrapLines(query, valuePrototye))
  return "Insert query yanked."
}
