package main

type EditorState struct {
  text Text
  cursorX, cursorY int
}

func NewEditorState(text Text, cX, cY int) *EditorState {
  return &EditorState{text, cX, cY}
}

func (es *EditorState) Unpack() (Text, int, int) {
  return es.text.Clone(), es.cursorX, es.cursorY
}

type DumbHistory struct {
  current int

  states []*EditorState
}

func NewDumbHistory(max int) *DumbHistory {
  dh := &DumbHistory{}
  maxSize := Max(1, max)
  dh.states = make([]*EditorState, maxSize)
  return dh
}

func (dh *DumbHistory) Add(b *EditorState) {
  if dh.current == len(dh.states) - 1 {
    if dh.states[dh.current] != nil {
      copy(dh.states, dh.states[1:])
    }
    dh.states[dh.current] = b

  } else {
    dh.states[dh.current] = b
    dh.current += 1

    for i := dh.current; i < len(dh.states); i++ {
      dh.states[i] = nil
    }
  }
}

func (dh *DumbHistory) Push(b *EditorState) {
  dh.RedoToLast()
  dh.Add(b)
}

func (dh *DumbHistory) RedoToLast() {
  for dh.current < len(dh.states) - 1 {
    if dh.states[dh.current] == nil {
      break
    }
    dh.current += 1
  }
}

func (dh *DumbHistory) Current() *EditorState {
  return dh.states[dh.current]
}

func (dh *DumbHistory) Undo() *EditorState {
  if dh.current == 0 {
    return nil
  }
  dh.current -= 1
  return dh.states[dh.current]
}

func (dh *DumbHistory) Redo() *EditorState {
  if dh.current == len(dh.states) - 1 || dh.states[dh.current] == nil {
    return nil
  }
  dh.current += 1
  return dh.states[dh.current]
}
