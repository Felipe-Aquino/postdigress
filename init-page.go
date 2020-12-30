package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
)

type InitPage struct {
	msg    *tview.TextView
  form   *tview.Form
  layout *tview.Grid
}

func NewInitPage(c *Context) *InitPage {
  msg := ""
  config, err := ReadConfigFile()

  if err != nil {
    config = &Config{[]Connection{}}
    if err.Error() == "json_error" {
      msg = "Invalid config file found."
    }
  }

  c.config = config

  info := DefaultDbInfo(c.config)

  ip := &InitPage{}

	ip.msg = tview.NewTextView().
    SetDynamicColors(true).
    SetRegions(true).
    SetTextAlign(tview.AlignCenter)

  ip.msg.SetText("\n[wheat]Ctrl-C to quit[white]\n" + msg)

  ip.form = tview.NewForm()

  ip.form.
		AddInputField("Host", info.host, 0, nil, nil).
		AddInputField("Port", info.port, 0, nil, nil).
		AddInputField("User", info.user, 0, nil, nil).
		AddInputField("Password", info.pass, 0, nil, nil).
		//AddPasswordField("Password", info.pass, 0, '*', nil).
		AddInputField("Database", info.name, 0, nil, nil).
		AddCheckbox("Enable SSL", info.ssl, nil).
		AddButton("Connect", func() {

      // Maybe create a proper treatment for the case in which
      // the user tries to connect again 
      if c.loading.waiting {
        return
      }

      c.info = &DBInfo{
        host: GetFormInputValue(ip.form, 0),
        port: GetFormInputValue(ip.form, 1),
        user: GetFormInputValue(ip.form, 2),
        pass: GetFormInputValue(ip.form, 3),
        name: GetFormInputValue(ip.form, 4),
        ssl:  GetFormCheckValue(ip.form, 5),
      }

      c.loading.SetTextView(ip.msg)
      go c.loading.Init(c.app)

      go func () {
        db, err := ConnectDb(c.info)

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
		AddButton("Save", func() {
      c.mainPages.SwitchToPage("Conn")
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

  handler := GetFormKeyHandler(c.app, ip.form, 6, 2)

  ip.form.
    SetInputCapture(handler)

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

func (ip *InitPage) UpdateForm(c *Connection) {
  SetFormInputValue(ip.form, 0, c.Host)
  SetFormInputValue(ip.form, 1, c.Port)
  SetFormInputValue(ip.form, 2, c.User)
  SetFormInputValue(ip.form, 3, c.Pass)
  SetFormInputValue(ip.form, 4, c.Db)
  SetFormCheckValue(ip.form, 5, c.Ssl)
}

func DefaultDbInfo(c *Config) *DBInfo {
  info := &DBInfo{ user: "postgres", host: "localhost", port: "5432" }

  for i := 0; i < len(c.Connections); i++ {
    if c.Connections[i].IsDefault {
      info.user = c.Connections[i].User
      info.name = c.Connections[i].Db
      info.host = c.Connections[i].Host
      info.port = c.Connections[i].Port
      info.pass = c.Connections[i].Pass
      info.ssl = c.Connections[i].Ssl
      break
    }
  }

  return info
}
