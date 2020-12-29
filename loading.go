package main

import (
	"github.com/rivo/tview"
  "fmt"
  "time"
)

type Loading struct {
  tv *tview.TextView
  pos int
  direction int

  waiting bool
  closed chan bool
}

func NewLoading(tv *tview.TextView) *Loading {
  return &Loading{
    tv: tv,
    pos: 0,
    waiting: false,
    closed: make(chan bool),
  }
}

func InsertTag(s string, at int) string {
  return s[:at] + "[\"0\"]" + s[at: at + 2] + "[\"\"]" + s[at + 2:] 
}

func (l *Loading) SetTextView(tv *tview.TextView) {
  l.tv = tv
}

func (l *Loading) Init(app *tview.Application) {
  if l.tv == nil {
    return
  }

  l.pos = 0
  l.direction = 0

  l.waiting = true

  ticker := time.NewTicker(time.Millisecond)
  lastTime := time.Now()
  startTime := time.Now()

  for l.waiting {
    currTime := <-ticker.C
    if currTime.Sub(lastTime) >= (150 * time.Millisecond) {
      if l.direction == 0 {
        l.pos += 1
        if l.pos > 9 {
          l.direction = 1
        }
      } else {
        l.pos -= 1
        if l.pos < 2 {
          l.direction = 0
        }
      }

      app.QueueUpdateDraw(func() {
        l.tv.Clear()
      })

      app.QueueUpdateDraw(func() {
        fmt.Fprint(l.tv, InsertTag("|  Loading  |", l.pos))
        l.tv.Highlight("0")
      })

      lastTime = currTime
    }

    if currTime.Sub(startTime) >= (5 * time.Minute) {
      app.QueueUpdateDraw(func() {
        l.tv.Clear()
      })

      app.QueueUpdateDraw(func() {
        fmt.Fprint(l.tv, "Waiting for too long! Something wrong might have happened!.")
      })

      l.Close()
    }
  }
  l.closed <- true
}

func (l *Loading) Close() {
  if l.waiting {
    l.waiting = false
    <-l.closed
  }
}
