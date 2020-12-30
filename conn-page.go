// Demo code for the Grid primitive.
package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "fmt"
)

type ConnPage struct {
  showingMsg bool
  msg *tview.TextView

  selector *Selector
  connections []Connection

  form   *tview.Form
  layout *tview.Grid
}

func NewConnPage(c *Context) *ConnPage {
  cp := &ConnPage{}

  cp.showingMsg = false

  backConn := Connection{ Name: "[ Go Back ]" }
  newConn  := Connection{ Name: "[ New + ]\n", Host: "localhost", Port: "5432" }

  cp.connections = append([]Connection{backConn, newConn}, c.config.Connections...)

  cp.mountLayout()

  cp.selector.SetSelectedCb(func (idx int) {
    if idx == 0 {
      c.mainPages.SwitchToPage("Init")
      return
    }

    conn := cp.connections[idx]

    conn.WriteToForm(cp.form)

    if idx == 1 {
      SetFormInputValue(cp.form, 0, "")
    }
    c.SetFocus(cp.form)
  })

  cp.selector.SetDeleteCb(func (idx int) bool {
    if idx < 2 {
      return false
    }

    conns, _ := Connections(cp.connections).Remove(idx).(Connections)
    cp.connections = []Connection(conns)
    c.config.Connections = cp.connections[2:]
    err := WriteConfigFile(c.config)

    if err != nil {
      return false
    }

    return true
  })

  cp.form.
    AddButton("Save", func () {
      idx := cp.selector.SelectedItemIndex()

      if idx == 1 {
        cp.connections = append(cp.connections, Connection{})
        idx = len(cp.connections) - 1
      }

      oldName := cp.connections[idx].Name
      cp.connections[idx].ReadFromForm(cp.form)

      if cp.connections[idx].Name == "" {
        cp.connections[idx].Name = fmt.Sprintf("Conn %d", idx - 1)
      }

      if cp.connections[idx].IsDefault {
        for i := 0; i < len(cp.connections); i++ {
          cp.connections[i].IsDefault = false

          if idx == i {
            cp.connections[i].IsDefault = true
            c.initPage.UpdateForm(&cp.connections[i])
          }
        }
      }

      c.config.Connections = cp.connections[2:]
      err := WriteConfigFile(c.config)

      if err == nil {
        (&Connection{}).WriteToForm(cp.form)
        cp.msg.SetText("[lightgreen]Connection saved in ~/.postdigress[white]")
        c.app.SetFocus(cp.selector.tv)

        if cp.connections[idx].Name != oldName {
          cp.selector.SetItems(Connections(cp.connections))
        } else {
          cp.selector.SelectItem(-1)
        }
      } else {
        cp.msg.SetText("[red]Error: " + err.Error() + "[white]")
      }
      cp.showingMsg = true
    }).
    AddButton("Cancel", func () {
      idx := cp.selector.SelectedItemIndex()
      conn := cp.connections[idx]
      conn.WriteToForm(cp.form)

      if idx == 1 {
        SetFormInputValue(cp.form, 0, "")
      }

      (&Connection{}).WriteToForm(cp.form)
      c.app.SetFocus(cp.selector.tv)
      cp.selector.SelectItem(-1)
    })

  handler := GetFormKeyHandler(c.app, cp.form, 8, 2)
  cp.form.
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      if cp.showingMsg {
        cp.msg.SetText("")
        cp.showingMsg = false
      }
      return handler(event)
    })

  return cp
}


func (cp *ConnPage) mountLayout() {
  cp.form = tview.NewForm()

  cp.msg = tview.NewTextView()
  cp.msg.
    SetTextAlign(tview.AlignCenter).
    SetDynamicColors(true).
    SetText("[wheat]D to delete a connection[white]")

  connTitle := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetText("Connections")

  menu := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)

  cp.selector = NewSelector(menu, Connections(cp.connections), true)

  cp.form.
    AddInputField("Conn. Name", "", 0, nil, nil).
    AddInputField("Host", "", 0, nil, nil).
    AddInputField("Port", "", 0, nil, nil).
    AddInputField("User", "", 0, nil, nil).
    AddInputField("Password", "", 0, nil, nil).
    AddInputField("Database", "", 0, nil, nil).
    AddCheckbox("Enable SSL", false, nil).
    AddCheckbox("Default", false, nil)

  cp.form.
    SetFieldBackgroundColor(tcell.NewRGBColor(100, 50, 200)).
    SetButtonBackgroundColor(tcell.NewRGBColor(100, 50, 200)).
    SetButtonsAlign(tview.AlignRight)

  flex := tview.NewFlex().
    SetDirection(tview.FlexRow).
    AddItem(cp.form, 0, 1, true).
    AddItem(cp.msg, 1, 1, false)

	cp.layout = tview.NewGrid().
    SetBorders(true).
		SetRows(-1, 1, -3, -1).
		SetColumns(-1, 20, -2, -1).
		AddItem(connTitle, 1, 1, 1, 1, 0, 0, false).
		AddItem(menu, 2, 1, 1, 1, 0, 0, true).
		AddItem(flex, 1, 2, 2, 1, 0, 0, false)
}

func (cp *ConnPage) Layout() tview.Primitive {
  return cp.layout
}

