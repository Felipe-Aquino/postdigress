package main

import (
  "strings"
  "errors"
  "time"
  "fmt"
)

type Command struct {
  list map[string]interface{}
}

func NewCommand() *Command {
  c := &Command{}
  c.list = make(map[string]interface{})

  c.Register("add",  CommandAdd)
  c.Register("sub",  CommandSub)
  c.Register("time", CommandTime)
  c.Register("utc", CommandToUTC)
  return c
}

func (c *Command) Register(name string, fn interface{}) {
  c.list[name] = fn
}

func (c *Command) Run(commandStr string) (string, error) {
  values := strings.Split(commandStr, " | ")
  values = SFilter(values, func(s string) bool { return len(s) > 0 })

  if len(values) > 1 {
    return c.Compose(values)
  }

  values = strings.Split(values[0], " ")
  values = SFilter(values, func(s string) bool { return len(s) > 0 })

  if len(values) > 0 {
    name   := values[0]
    params := values[1:]

    fn := c.list[name]

    if fn != nil {
      return CallFunction(fn, params)
    }
  }

  return "", errors.New("Fail to run command or command doens't exists.")
}

func (c *Command) Compose(commands []string) (string, error) {
  result := ""
  for _, command := range commands {
    _command := command + " " + result 

    v, err := c.Run(_command)

    if err != nil {
      return "", err
    }

    result = v
  }

  return result, nil
}

func CommandAdd(a, b float64) float64 {
  return a + b
}

func CommandSub(a, b float64) float64 {
  return a - b
}

func CommandTime() string {
  now := time.Now()
  return fmt.Sprint(now.Format(time.RFC3339))
}

func CommandToUTC(s string) string {
  tm, err := time.Parse(time.RFC3339, s)

  if err != nil {
    return fmt.Sprint(tm.UTC().Format(time.RFC3339))
  }

  return err.Error()
}
