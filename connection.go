package main

import (
	"github.com/rivo/tview"
  "encoding/json"
  "os/user"
	"io/ioutil"
  "errors"
)

type Connection struct {
  Name string `json:"name"`
  Host string `json:"host"`
  Port string `json:"port"`
  User string `json:"user"`
  Pass string `json:"pass"`
  Db   string `json:"db"`
  Ssl  bool   `json:"ssl"`
  IsDefault  bool `json:"default"`
}

type Connections []Connection

func (c Connections ) GetItemName(index int) string {
  if index < len(c) {
    return c[index].Name
  }
  return ""
}

func (c Connections) Len() int {
  return len(c)
}

func (c Connections) Remove(at int) Enumerable {
  connLen := len(c)

  if at == connLen - 1 {
    return c[:at]
  }

  if connLen  > 0 && at < connLen  - 1 {
    return append(c[:at], c[at + 1:]...)
  }

  return Connections([]Connection{})
}

func (c* Connection) ReadFromForm(form *tview.Form) {
  c.Name = GetFormInputValue(form, 0)
  c.Host = GetFormInputValue(form, 1)
  c.Port = GetFormInputValue(form, 2)
  c.User = GetFormInputValue(form, 3)
  c.Pass = GetFormInputValue(form, 4)
  c.Db   = GetFormInputValue(form, 5)
  c.Ssl  = GetFormCheckValue(form, 6)
  c.IsDefault = GetFormCheckValue(form, 7)
}

func (c* Connection) WriteToForm(form *tview.Form) {
  SetFormInputValue(form, 0, c.Name)
  SetFormInputValue(form, 1, c.Host)
  SetFormInputValue(form, 2, c.Port)
  SetFormInputValue(form, 3, c.User)
  SetFormInputValue(form, 4, c.Pass)
  SetFormInputValue(form, 5, c.Db)
  SetFormCheckValue(form, 6, c.Ssl)
  SetFormCheckValue(form, 7, c.IsDefault)
}

type Config struct {
  Connections []Connection `json:"connections"`
}

func ReadConfigFile() (*Config, error) {
  usr, err := user.Current()
  dir := usr.HomeDir

  path := dir + "/.postdigress"

  if err != nil {
    return nil, errors.New("invalid_path")
  }

  file, err := ioutil.ReadFile(path)

  if err != nil {
    return nil, errors.New("read_error")
  }

  config := &Config{}

	err = json.Unmarshal([]byte(file), config)

  if err != nil {
    return nil, errors.New("json_error")
  }

  return config, nil
}

func WriteConfigFile(config *Config) error {
  usr, err := user.Current()
  dir := usr.HomeDir

  path := dir + "/.postdigress"

  if err != nil {
    return errors.New("invalid_path")
  }

  data, err := json.Marshal(*config)

  if err != nil {
    return err
  }

	err = ioutil.WriteFile(path, data, 0644)

  return err
}

