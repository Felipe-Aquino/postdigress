package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "fmt"
  "sort"
)

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

func main() {
	app := tview.NewApplication()
	menuBar := tview.NewTextView()

	mainPages := tview.NewPages()
	sqlPages := tview.NewPages()

  context := &Context{
    app: app,
    mainPages: mainPages,
    menuBar: menuBar,
    loading: NewLoading(nil),
  }

  initPage   := NewInitPage(context)
  runPage    := NewRunPage(context)
  structPage := NewStructPage(context)

  context.runPage = runPage
  context.structPage = structPage

	mainPages.AddPage("Init", initPage.Layout(), true, true)

  sqlPages.
    AddPage("0", runPage.Layout(), true, false).
    AddPage("1", structPage.Layout(), true, false)

	menuBar.
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetHighlightedFunc(func(added, removed, remaining []string) {
			sqlPages.SwitchToPage(added[0])
      context.FocusMenu()

      if added[0] == "1" {
        context.loading.SetTextView(structPage.dbSelect)
        go context.loading.Init(context.app)

        go func () {
          result := GetTables(context.db)
          if result.err != nil {
            structPage.dbSelect.SetText(result.err.Error())
            context.loading.Close()
            return
          }

          context.loading.Close()
          tables := SSMap(result.values, func (arr []string) string {
            return arr[0]
          })
          sort.Strings(tables)

          context.Enqueue(func () {
            structPage.SetDBTitleName(context.info.name)
            structPage.SetTables(tables)
          })

        }()
      }

		})

  fmt.Fprint(menuBar, ` ["0"][yellow] [::bu]E[::-]xecute [white][""] | ["1"][yellow] [::bu]S[::-]tructure [white][""] | ["2"][yellow] [::bu]Q[::-]uit [white][""]`)

  menuBar.Highlight("0")
  context.selectedMenu = RUN_MENU

	mainPages.AddPage("SQL", sqlPages, true, false)

	if err := app.SetRoot(mainPages, true).Run(); err != nil {
		panic(err)
	}

}
