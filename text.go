package main

type Text []string

func WrapLines(lines ...string) Text {
  return Text(lines)
}

func (t Text) Len() int {
  return len(t)
}

func (t Text) Clone() Text {
  return Text(append([]string{}, t...))
}

func (t Text) FirstLine() string {
  if len(t) > 0 {
    return t[0]
  }
  return ""
}

func (t Text) LastLine() string {
  if len(t) > 0 {
    return t[len(t) - 1]
  }
  return ""
}

func (t Text) Line(at int) string {
  return t[at]
}

func (t Text) IsLineEmpty(l int) bool {
  if l >= 0 && l < len(t) {
    return len(t[l]) == 0
  }
  return true
}

func (t Text) LineLen(l int) int {
  if l < len(t) {
    return len(t[l])
  }
  return -1
}

func (t Text) ReplaceChar(i, j int, ch rune) {
  t[i] = t[i][:j] + string(ch) + t[i][j + 1:]
}

func (t Text) SubStrAt(line, colStart, colEnd int) Text {
  if line < len(t) && colStart <= colEnd && colEnd <= len(t[line]) {
    return WrapLines( t[line][colStart: colEnd] )
  }
  return WrapLines("")
}

func (t Text) DeleteSubStrAt(l, colStart, colEnd int) Text {
  if l < len(t) && colStart <= colEnd && colEnd <= len(t[l]) {
    substr := t[l][colStart: colEnd]
    t[l] = t[l][:colStart] + t[l][colEnd:]
    return WrapLines(substr)
  }
  return WrapLines("")
}

func (t Text) SubText(start, size int) Text {
  if start < len(t) {
    end := Min(len(t), start + size)
    return Text(t[start: end])
  }
  return Text([]string{})
}

func (t Text) InsertText(start int, lines Text) Text {
  if start <= len(t) && len(t) > 0 {
    return Text(append(t[:start], append(lines, t[start:]...)...))
  }
  return Text(append(t, lines...))
}

// From a start line, inserts line by line into the orignal text
func (t Text) InsertLines(start int, lines ...string) Text {
  if start <= len(t) && len(t) > 0 {
    return Text(append(t[:start], append(lines, t[start:]...)...))
  }
  return Text(append(t, lines...))
}

func (t Text) InsertAt(l, c int, value Text) Text {
  if len(value) == 0 {
    return t
  }

  if len(t) == 0{
    return value
  }

  if l >= 0 && l < len(t) {
    result := t

    line := t[l]

    if value.Len() == 1 {
      result[l] = line[:c] + value[0] + line[c:]

    } else if value.Len() > 1 {
      lineStart := line[:c] + value.FirstLine()
      lineEnd   := value.LastLine() + line[c:]

      size := value.Len() - 1
      result = t.InsertText(l, value.SubText(0, size)) 

      result[l]            = lineStart
      result[l + size] = lineEnd
    }

    return result
  }

  return Text([]string{})
}

// Delete text lines from a start line to an end line
func (t Text) DeleteLines(start, end int) Text {
  if start <= end && end < len(t) {
    return Text(append(t[:start], t[end + 1:]...))
  }
  return t
}

func (t Text) DeleteRange(lineStart, colStart, lineEnd, colEnd int) Text {
  if len(t) == 0 || lineStart > lineEnd {
    return t
  }

  if lineStart < len(t) && colStart <= t.LineLen(lineStart) && colEnd < t.LineLen(lineEnd) {
    t[lineStart] = t[lineStart][:colStart] + t[lineEnd][colEnd + 1:]

    return t.DeleteLines(lineStart + 1, lineEnd)
  }

  return t
}

func (t Text) CopyRange(lineStart, colStart, lineEnd, colEnd int) Text {
  if len(t) == 0 || lineStart > lineEnd {
    return Text([]string{})
  }

  if colStart <= t.LineLen(lineStart) && colEnd < t.LineLen(lineEnd) {
    if lineStart < len(t) {
      if lineStart == lineEnd {
        return WrapLines(t[lineStart][colStart: colEnd + 1])
      }

      result := []string{t[lineStart][colStart:]}
      result = append(result, t[lineStart + 1: lineEnd]...)
      result = append(result, t[lineEnd][:colEnd + 1])
      return result
    }
  }

  return Text([]string{})
}