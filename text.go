package main

import "strings"

type Line []rune

func (l Line) String() string {
  return string(l)
}

func (l Line) Clone() Line {
  newLine := make([]rune, len(l))
  copy(newLine, l)
  return Line(newLine)
}

func LinesFromStrs(strs []string) []Line {
  lines := []Line{}

  for _, str := range strs {
    lines = append(lines, Line(str))
  }
  return lines
}

type Text []Line

func WrapLines(lines ...string) Text {
  return Text(LinesFromStrs(lines))
}

func WrapLinesR(lines ...Line) Text {
  return Text(lines)
}

func (t Text) Len() int {
  return len(t)
}

func (t Text) Clone() Text {
  newText := Text([]Line{})

  for _, l := range t {
     newText = append(newText, l.Clone())
  }

  return newText;
}

func (t Text) FirstLine() Line {
  if len(t) > 0 {
    return t[0]
  }
  return Line{}
}

func (t Text) LastLine() Line {
  if len(t) > 0 {
    return t[len(t) - 1]
  }
  return Line{}
}

func (t Text) Line(at int) Line {
  return t[at]
}

func (t Text) SetLine(at int, value Line) {
  if at < t.Len() {
    t[at] = value
  }
}

func (t Text) LineLen(l int) int {
  if l < len(t) {
    return len(t[l])
  }
  return -1
}

func (t Text) ReplaceChar(i, j int, ch rune) {
  t[i][j] = ch
}

func (t Text) SubStrAt(line, colStart, colEnd int) Text {
  if line < len(t) && colStart <= colEnd && colEnd <= len(t[line]) {
    return WrapLinesR(t[line][colStart: colEnd])
  }
  return WrapLines("")
}

func (t Text) DeleteSubStrAt(l, colStart, colEnd int) Text {
  if l < len(t) && colStart <= colEnd && colEnd <= len(t[l]) {
    substr := t[l][colStart: colEnd]
    t[l] = append(t[l][:colStart], t[l][colEnd:]...)
    return WrapLinesR(substr)
  }
  return WrapLines("")
}

func (t Text) SubText(start, size int) Text {
  if start < len(t) {
    end := Min(len(t), start + size)
    return Text(t[start: end])
  }
  return Text([]Line{})
}

func (t Text) InsertText(start int, lines Text) Text {
  if start <= len(t) && len(t) > 0 {
    return Text(append(t[:start], append(lines, t[start:]...)...))
  }
  return Text(append(t, lines...))
}

// From a start line, inserts line by line into the orignal text
func (t Text) InsertLines(start int, lines ...string) Text {
  linesR := WrapLines(lines...)
  if start <= len(t) && len(t) > 0 {
    return Text(append(t[:start], append(linesR, t[start:]...)...))
  }
  return Text(append(t, linesR...))
}

func (t Text) InsertAt(l, c int, value Text) Text {
  if len(value) == 0 {
    return t
  }

  if len(t) == 0 {
    return value
  }

  if l >= 0 && l < len(t) {
    result := t

    line := t[l]

    if value.Len() == 1 {
      result[l] = append(line[:c], append(value[0], line[c:]...)...)

    } else if value.Len() > 1 {
      lineStart := append(line[:c], value.FirstLine()...)
      lineEnd   := append(value.LastLine(), line[c:]...)

      size := value.Len() - 1
      result = t.InsertText(l, value.SubText(0, size)) 

      result[l]        = lineStart
      result[l + size] = lineEnd
    }

    return result
  }

  return Text([]Line{})
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
    t[lineStart] = append(t[lineStart][:colStart], t[lineEnd][colEnd + 1:]...)

    return t.DeleteLines(lineStart + 1, lineEnd)
  }

  return t
}

func (t Text) CopyRange(lineStart, colStart, lineEnd, colEnd int) Text {
  if len(t) == 0 || lineStart > lineEnd {
    return Text([]Line{})
  }

  if colStart <= t.LineLen(lineStart) && colEnd < t.LineLen(lineEnd) {
    if lineStart < len(t) {
      if lineStart == lineEnd {
        return WrapLinesR(t[lineStart][colStart: colEnd + 1])
      }

      result := []Line{t[lineStart][colStart:]}
      result = append(result, t[lineStart + 1: lineEnd]...)
      result = append(result, t[lineEnd][:colEnd + 1])
      return result
    }
  }

  return Text([]Line{})
}

func (t Text) String() string {
  var builder strings.Builder

  for _, line := range t {
    builder.WriteString(line.String())
    builder.WriteString("\n")
  }
  return builder.String()
}

func TextFromString(s string) Text {
  return WrapLines(strings.Split(s, "\n")...)
}
