package main

import (
	"github.com/rivo/tview"
  "github.com/gdamore/tcell"
  "fmt"
  "hash/fnv"
  "reflect"
  "strconv"
  "errors"
  "os/user"
	"io/ioutil"
)

func IsAlpha(r rune) bool {
  return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func IsAlnum(r rune) bool {
  return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func IsDigit(r rune) bool {
  return r >= '0' && r <= '9'
}

func IsSpace(r rune) bool {
  return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func IsSpecialChar(r rune) bool {
  return !IsAlnum(r) && !IsSpace(r) && r != '_'
}

func IsWordChar(r rune) bool {
  return IsAlnum(r) || r == '_'
}

func Tern(cond bool, v1, v2 int) int {
  if cond {
    return v1
  }
  return v2
}

func Hash(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}

func Min(a, b int) int {
  if a > b {
    return b
  }
  return a
}

func Max(a, b int) int {
  if a < b {
    return b
  }
  return a
}

func StrClip(s string, max int) string {
  if len(s) > max {
    return s[:max]
  }
  return s
}

func StrClamp(s string, min, max int) string {
  size := len(s)
  if size > min {
    if size > max {
      return s[min:max]
    }
    return s[min:]
  }
  return ""
}

func SSMap(vs [][]string, f func([]string) string) []string {
  vsm := make([]string, len(vs))
  for i, v := range vs {
    vsm[i] = f(v)
  }
  return vsm
}

func SFilter(vs []string, f func(string) bool) []string {
  filtered := []string{} 
  for _, v := range vs {
    if f(v) {
      filtered = append(filtered, v)
    }
  }
  return filtered
}

func ExpandHomeDir(path string) (string, error) {
  if path[0] == '~' {
    usr, err := user.Current()

    if err != nil {
      return "", errors.New("Invalid path.")
    }

    dir := usr.HomeDir
    path = dir + path[1:]
  }
  return path, nil
}

func ReadFile(path string) (string, error) {
  path, err := ExpandHomeDir(path)

  if err != nil {
    return "", err
  }

  file, err := ioutil.ReadFile(path)

  if err != nil {
    return "", errors.New("Read error")
  }

  return string(file), nil
}

func WriteFile(path string, data string) error {
  path, err := ExpandHomeDir(path)

  if err != nil {
    return err
  }

	err = ioutil.WriteFile(path, []byte(data), 0644)

  if err != nil {
    return errors.New("Read error")
  }

  return nil
}

func CallFunction(f interface{}, values []string) (string, error) {
	tf := reflect.TypeOf(f)
	if tf.Kind() != reflect.Func {
		return "", errors.New("expects a function")
	}

	vf := reflect.ValueOf(f)

  numParams := tf.NumIn()

  if len(values) < numParams {
		return "", errors.New("Mismatch between param numbers.")
  }

  params := []reflect.Value{}

  for i := 0; i < numParams; i++ {
    param := reflect.ValueOf("")

    paramType := tf.In(i)

    switch paramType.Kind() {
    case reflect.String:
      param = reflect.ValueOf(values[i])
    case reflect.Int:
      n, err := strconv.ParseInt(values[i], 10, 32)
      if err == nil {
        param = reflect.ValueOf(int(n))
      } else {
        param = reflect.ValueOf(int(0))
      }
    case reflect.Int32:
      n, err := strconv.ParseInt(values[i], 10, 32)
      if err == nil {
        param = reflect.ValueOf(int32(n))
      } else {
        param = reflect.ValueOf(int32(0))
      }
    case reflect.Int64:
      n, err := strconv.ParseInt(values[i], 10, 64)
      if err == nil {
        param = reflect.ValueOf(n)
      } else {
        param = reflect.ValueOf(int64(0))
      }
    case reflect.Float32:
      n, err := strconv.ParseFloat(values[i], 32)
      if err == nil {
        param = reflect.ValueOf(n)
      } else {
        param = reflect.ValueOf(float32(0.0))
      }
    case reflect.Float64:
      n, err := strconv.ParseFloat(values[i], 64)
      if err == nil {
        param = reflect.ValueOf(n)
      } else {
        param = reflect.ValueOf(float64(0.0))
      }
    case reflect.Bool:
      if values[i] == "true" {
        param = reflect.ValueOf(true)
      } else {
        param = reflect.ValueOf(false)
      }
    default:
		  return "", errors.New("Unexpected param type")
    }

    params = append(params, param)
  }

  out := vf.Call(params)

  returnValue := ""

  if len(out) > 0 {
    returnValue = fmt.Sprint(out[0])
  }

  return returnValue, nil
}

///* tview *///

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

func GetFormInputValue(form *tview.Form, index int) string {
  field, ok := form.GetFormItem(index).(*tview.InputField)
  if ok {
    return field.GetText()
  }
  return ""
}

func GetFormCheckValue(form *tview.Form, index int) bool {
  check, ok := form.GetFormItem(index).(*tview.Checkbox)
  if ok {
    return check.IsChecked()
  }
  return false
}

func SetFormInputValue(form *tview.Form, index int, value string) {
  field, ok := form.GetFormItem(index).(*tview.InputField)
  if ok {
    field.SetText(value)
  }
}

func SetFormCheckValue(form *tview.Form, index int, value bool) {
  check, ok := form.GetFormItem(index).(*tview.Checkbox)
  if ok {
    check.SetChecked(value)
  }
}

func GetFormKeyHandler(
  app *tview.Application,
  form *tview.Form, items,
  buttons int) func (event *tcell.EventKey) *tcell.EventKey {

  return func (event *tcell.EventKey) *tcell.EventKey {
    idx, btn := form.GetFocusedItemIndex()

    if btn == -1 {
      switch event.Key() {
      case tcell.KeyDown, tcell.KeyCtrlJ, tcell.KeyTab, tcell.KeyCtrlN:
        idx = idx + 1
        if idx < items {
          item := form.GetFormItem(idx)
          app.SetFocus(item)
        } else {
          item := form.GetButton(0)
          app.SetFocus(item)
        }

        return nil
      case tcell.KeyUp, tcell.KeyCtrlK, tcell.KeyCtrlP:
        idx--
        if idx < 0 {
          item := form.GetButton(0)
          app.SetFocus(item)
          return nil
        }
        item := form.GetFormItem(idx)
        app.SetFocus(item)
      }
    } else {
      switch event.Key() {
      case tcell.KeyDown, tcell.KeyCtrlJ, tcell.KeyCtrlN:
        item := form.GetFormItem(0)
        app.SetFocus(item)
      case tcell.KeyUp, tcell.KeyCtrlK, tcell.KeyCtrlP:
        item := form.GetFormItem(items - 1)
        app.SetFocus(item)
      case tcell.KeyLeft, tcell.KeyCtrlH, tcell.KeyRight, tcell.KeyCtrlL:
        item := form.GetButton((btn + 1) % buttons)
        app.SetFocus(item)
      case tcell.KeyTab:
        if btn == 0 {
          item := form.GetButton((btn + 1) % buttons)
          app.SetFocus(item)
        } else {
          item := form.GetFormItem(0)
          app.SetFocus(item)
        }
        return nil
      }
    }
    return event
  }
}

