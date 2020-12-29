package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "fmt"
  "strings"
  "regexp"
  "strconv"
)

var columnsFields [6]string = [6]string{"Column", "Type", "Default", "Primary Key", "Null?"}

var indexesFields [6]string = [6]string{"Name", "P. key", "Unique", "Column", "Size"}

type StructPage struct {
  dbTitle      *tview.TextView
  columnsTitle *tview.TextView
  indexesTitle *tview.TextView

  dbSelect *tview.TextView

  columnsTable *tview.Table
  indexesTable *tview.Table

  layout *tview.Grid

  focusedType ComponentType

  tables []string
  selectedTable int
  cursor int
}

func NewStructPage(c *Context) *StructPage {
  sp := &StructPage{}

  sp.focusedType = MENU

  sp.selectedTable = -1
  sp.tables = []string{}

  sp.dbTitle = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

  sp.dbTitle.SetText(" [\"0\"] [::bu]D[::-]B: ??? [\"\"]")

  sp.dbSelect = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

  sp.dbSelect.
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      if event.Rune() == 'j' {
        if sp.cursor < len(sp.tables) - 1 {
          sp.cursor += 1
          sp.dbSelect.Highlight(strconv.Itoa(sp.cursor))
        }
      } else if event.Rune() == 'k' {
        if sp.cursor > 0 {
          sp.cursor -= 1
          sp.dbSelect.Highlight(strconv.Itoa(sp.cursor))
        }
      } else if event.Key() == tcell.KeyCR {
        sp.SelectTable(sp.cursor)
        go sp.QueryTableInfo(c, sp.tables[sp.selectedTable])
      }

      return event
    })

  sp.SetTables([]string{})

  sp.columnsTable = tview.NewTable().
		SetBorders(false).
		SetSeparator(tview.Borders.Vertical)

  sp.indexesTable = tview.NewTable().
		SetBorders(false).
		SetSeparator(tview.Borders.Vertical)

  TableSetData(sp.columnsTable, columnsFields[:], [][]string{}, false)

  sp.columnsTitle = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
    SetText(" [\"1\"] [::bu]C[::-]olumns [\"\"]")

  TableSetData(sp.indexesTable, indexesFields[:], [][]string{}, false)

  sp.indexesTitle = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
    SetText(" [\"2\"] [::bu]I[::-]ndexes [\"\"]")

  sp.layout = tview.NewGrid().
		SetBorders(true).
		SetRows(1, 1, -2, 1, -1).
		SetColumns(-1, -3).
		AddItem(c.menuBar,       0, 0, 1, 4, 0, 0, false).
		AddItem(sp.dbTitle,      1, 0, 1, 1, 0, 0, false).
		AddItem(sp.dbSelect,     2, 0, 3, 1, 0, 0, false).
		AddItem(sp.columnsTitle, 1, 1, 1, 3, 0, 0, true).
		AddItem(sp.columnsTable, 2, 1, 1, 3, 0, 0, true).
		AddItem(sp.indexesTitle, 3, 1, 1, 3, 0, 0, false).
		AddItem(sp.indexesTable, 4, 1, 1, 3, 0, 0, false)

  sp.layout.
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      if sp.focusedType != DATABASE && event.Rune() == 'd' {
        sp.SetCompType(DATABASE)
        c.SetFocus(sp.dbSelect)

      } else if sp.focusedType != COLUMNS && event.Rune() == 'c' {
        sp.SetCompType(COLUMNS)
        c.SetFocus(sp.columnsTable)

      } else if sp.focusedType != INDEXES && event.Rune() == 'i' {
        sp.SetCompType(INDEXES)
        c.SetFocus(sp.indexesTable)

      } else if sp.focusedType == MENU {
        c.HandleMenuKeyInput(event)

      } else if event.Rune() == 'q' {
        sp.SetCompType(MENU)
        c.FocusMenu()
      }

      return event
    })

  return sp
}

func (sp *StructPage) SetDBTitleName(name string) {
  name = fmt.Sprintf(" [\"0\"] [::bu]D[::-]B: %s [\"\"]", name)
  sp.dbTitle.SetText(name)
}

func (sp *StructPage) SetTables(tables []string) {
  sp.tables = tables
  sp.selectedTable = -1
  sp.cursor = 0

  text := ""
  for i, table := range tables {
    text += fmt.Sprintf(" [\"%d\"] %s [\"\"]\n", i, table)
  }

  sp.dbSelect.SetText(text)
  sp.dbSelect.Highlight("0")
}

func (sp *StructPage) SelectTable(index int) {
  sp.selectedTable = index

  text := ""
  for i, table := range sp.tables {
    if i != index {
      text += fmt.Sprintf(" [\"%d\"] %s [\"\"]\n", i, table)
    } else {
      text += fmt.Sprintf(" [\"%d\"][orangered] %s [white][\"\"]\n", i, table)
    }
  }

  sp.dbSelect.SetText(text)
}

func (sp *StructPage) SetCompType(t ComponentType) {
  sp.focusedType = t

  switch t {
  case DATABASE:
    sp.dbTitle.Highlight("0")
    sp.columnsTitle.Highlight("")
    sp.indexesTitle.Highlight("")

  case COLUMNS:
    sp.dbTitle.Highlight("")
    sp.columnsTitle.Highlight("1")
    sp.indexesTitle.Highlight("")

  case INDEXES:
    sp.dbTitle.Highlight("")
    sp.columnsTitle.Highlight("")
    sp.indexesTitle.Highlight("2")

  case MENU:
    sp.dbTitle.Highlight("")
    sp.columnsTitle.Highlight("")
    sp.indexesTitle.Highlight("")
  }
}

func (sp *StructPage) Layout() tview.Primitive {
  return sp.layout
}

func (sp *StructPage) QueryTableInfo(c *Context, tablename string) {
  c.Enqueue(func () {
    TableSetData(sp.columnsTable, columnsFields[:], [][]string{}, false)
    TableSetData(sp.indexesTable, indexesFields[:], [][]string{}, false)
  })

  oid, err := GetOID(c.db, tablename)

  if err != nil {
    return
  }

  result := GetConstraints(c.db, oid)
  AdaptConstraintData(result.values)

  indexes := result.values

  result = GetColumns(c.db, tablename)
  result.values = ExtractColumnsData(result.values)

  columns := result.values

  c.Enqueue(func () {
    TableSetData(sp.columnsTable, columnsFields[:], columns, false)
    TableSetData(sp.indexesTable, indexesFields[:], indexes, false)
  })
}

func ExtractColumnsData(values [][]string) [][]string {
  result := make([][]string, len(values))

  for i, row := range values {
    result[i] = make([]string, 5)

    result[i][0] = row[0]

    switch row[1] {
    case "numeric":
      result[i][1] = fmt.Sprintf("numeric(%s, %s)", row[3], row[4])
    case "character varying":
      result[i][1] = fmt.Sprintf("varchar(%s)", row[2])
    case "character":
      result[i][1] = fmt.Sprintf("char(%s)", row[2])
    case "timestamp with time zone":
      result[i][1] = "timestamp w/ tz"
    default:
      result[i][1] = row[1]
    }

    if row[5] == "nil" {
      result[i][2] = ""
    } else if strings.HasPrefix(row[5], "nextval") {
      result[i][2] = "auto"
    } else {
      result[i][2] = row[5]
    }

    if row[7] == "true" {
      result[i][3] = "YES"
    } else {
      result[i][3] = ""
    }
    result[i][4] = row[6]
  }

  return result
}

func AdaptConstraintData(values [][]string) {
  re, err := regexp.Compile("btree \\((.*)\\)")

  if err != nil {
    fmt.Println("re", err.Error())
    return
  }

  for i := 0; i < len(values); i++ {
    matches := re.FindStringSubmatch(values[i][3])
    if len(matches) > 1 {
      values[i][3] = matches[1]
    }
  }
}

