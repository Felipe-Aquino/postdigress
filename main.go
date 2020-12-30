package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "fmt"
  "sort"
)

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

  context.initPage = initPage
  context.runPage = runPage
  context.structPage = structPage

  connPage := NewConnPage(context)

	mainPages.AddPage("Init", initPage.Layout(), true, true)
	mainPages.AddPage("Conn", connPage.Layout(), true, false)

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

  app.SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyCtrlC {
      context.Finish()
      return nil
    }
    return event
  })

	if err := app.SetRoot(mainPages, true).Run(); err != nil {
		panic(err)
	}

}
