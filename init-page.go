package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  //"time"
)

func FormTextValue(form *tview.Form, index int) string {
  field, ok := form.GetFormItem(index).(*tview.InputField)
  if ok {
    return field.GetText()
  }
  return ""
}

func FormCheckValue(form *tview.Form, index int) bool {
  check, ok := form.GetFormItem(index).(*tview.Checkbox)
  if ok {
    return check.IsChecked()
  }
  return false
}

type InitPage struct {
	msg    *tview.TextView
  form   *tview.Form
  layout *tview.Grid
}

func NewInitPage(c *Context) *InitPage {
  ip := &InitPage{}

  var info *DBInfo = nil

  if c.info != nil {
    info = c.info
  } else {
    info = &DBInfo{ "postgres", "postgres", "takeat", "localhost", "5432", false }
  }

  ip.form = tview.NewForm()

  ip.form.
		AddInputField("Host", info.host, 0, nil, nil).
		AddInputField("Port", info.port, 0, nil, nil).
		AddInputField("User", info.user, 0, nil, nil).
		//AddInputField("Password", info.pass, 0, nil, nil).
		AddPasswordField("Password", info.pass, 0, '*', nil).
		AddInputField("Database", info.name, 0, nil, nil).
		AddCheckbox("Enable SSL", info.ssl, nil).
		AddButton("Connect", func() {

      // Maybe create a proper treatment for the case in which
      // the user tries to connect again 
      if c.loading.waiting {
        return
      }

      c.info = &DBInfo{
        host: FormTextValue(ip.form, 0),
        port: FormTextValue(ip.form, 1),
        user: FormTextValue(ip.form, 2),
        pass: FormTextValue(ip.form, 3),
        name: FormTextValue(ip.form, 4),
        ssl: FormCheckValue(ip.form, 5),
      }

      c.loading.SetTextView(ip.msg)
      go c.loading.Init(c.app)

      go func () {
        //time.Sleep(10 * time.Second)
        db, err := ConnectDb(c.info.user, c.info.pass, c.info.name)

        if err != nil {
          c.loading.Close()
          c.Enqueue(func () {
            ip.msg.SetText(err.Error())
          })
          return
        }

        c.loading.Close()

        c.db = db
        c.Enqueue(func () {
          c.mainPages.SwitchToPage("SQL")
        })

      }()
    }).
		AddButton("Quit", func() {
			c.Finish()
		})

  ip.form.
    SetFieldBackgroundColor(tcell.NewRGBColor(100, 50, 200)).
    SetButtonBackgroundColor(tcell.NewRGBColor(100, 50, 200)).
    SetButtonsAlign(tview.AlignRight)

	ip.form.
    SetBorder(true).
    SetBorderPadding(2, 2, 2, 2).
    SetTitle(" Connection ").
    SetTitleAlign(tview.AlignLeft)

  ip.form.
    SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
      idx, btn := ip.form.GetFocusedItemIndex()

      if btn == -1 {
        if event.Key() == tcell.KeyDown ||
           event.Key() == tcell.KeyCtrlJ ||
           event.Key() == tcell.KeyTab {

          idx = idx + 1
          if idx < 6 {
            item := ip.form.GetFormItem(idx)
            c.app.SetFocus(item)
          } else {
            item := ip.form.GetButton(0)
            c.app.SetFocus(item)
            return nil
          }

          return nil

        } else if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyCtrlK {
          idx--
          if idx < 0 {
            item := ip.form.GetButton(0)
            c.app.SetFocus(item)
            return nil
          }
          item := ip.form.GetFormItem(idx)
          c.app.SetFocus(item)
        }

      } else {
        if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyCtrlJ {
          item := ip.form.GetFormItem(0)
          c.app.SetFocus(item)

        } else if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyCtrlK {
          item := ip.form.GetFormItem(5)
          c.app.SetFocus(item)

        } else if event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyCtrlH {
          item := ip.form.GetButton((btn + 1) % 2)
          c.app.SetFocus(item)

        } else if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyCtrlL {
          item := ip.form.GetButton((btn + 1) % 2)
          c.app.SetFocus(item)

        } else if event.Key() == tcell.KeyTab {
          if btn == 0 {
            item := ip.form.GetButton((btn + 1) % 2)
            c.app.SetFocus(item)
          } else {
            item := ip.form.GetFormItem(0)
            c.app.SetFocus(item)
          }

          return nil
        }
      }
      return event
    })

	ip.msg = tview.NewTextView().SetRegions(true)

	ip.layout = tview.NewGrid().
		SetBorders(false).
		SetRows(-1, -2, -1).
		SetColumns(-1, -1, -1).
		AddItem(ip.form, 1, 1, 1, 1, 0, 0, true).
		AddItem(ip.msg,  2, 1, 1, 1, 0, 0, false)

  return ip
}

func (ip *InitPage) Layout() tview.Primitive {
  return ip.layout
}

