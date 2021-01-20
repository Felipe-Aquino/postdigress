package main

type Validator func(rune) bool

func StrReadWhileBackwards(line Line, start int, isValid Validator) int {
  if start < 0 {
    return -1
  }

  c := line[start]
  j := start

  if isValid(c) {
    for  {
      j--
      if j < 0 {
        return -1
      }
      c := line[j]

      if !isValid(c) {
        return j
      }
    }
  }

  return start
}

func StrReadWhileForward(line Line, start int, isValid Validator) int {
  textLen := len(line)

  if start >= textLen {
    return textLen
  }

  c := line[start]
  j := start

  if isValid(c) {
    for  {
      j++
      if j > textLen - 1 {
        return j
      }
      c := line[j]

      if !isValid(c) {
        return j
      }
    }
  }

  return start
}

func FindNextWordStart(text Text, i, j int) (int, int, bool) {
  found := false

  for {
    lineLen := len(text[i])

    if lineLen == 0 {
      if i < len(text) - 1 {
        i, j = i + 1, 0
        if len(text[i]) > 0 && !IsSpace(text[i][j]) {
          found = true
          break
        }
        continue
      }

      break
    }

    if IsSpecialChar(text[i][j]) {
      j = StrReadWhileForward(text[i], j, IsSpecialChar)
    } else if IsWordChar(text[i][j]) {
      j = StrReadWhileForward(text[i], j, IsWordChar)
    }

    j = StrReadWhileForward(text[i], j, IsSpace)

    if j > lineLen - 1 {
      if i < len(text) - 1 {
        i, j = i + 1, 0
        if len(text[i]) > 0 && !IsSpace(text[i][j]) {
          found = true
          break
        }
        continue
      }

      break
    }

    found = true
    break
  }

  return i, j, found
}

func FindPrevWordEnd(text Text, i, j int) (int, int, bool) {
  found := false

  for {
    lineLen := len(text[i])

    if lineLen == 0 {
      if i > 0 {
        i, j = i - 1, len(text[i - 1]) - 1
        if j > 0 && !IsSpace(text[i][j]) {
          found = true
          break
        }
        continue
      }
      break
    }

    if IsSpecialChar(text[i][j]) {
      j = StrReadWhileBackwards(text[i], j, IsSpecialChar)
    } else if IsWordChar(text[i][j]) {
      j = StrReadWhileBackwards(text[i], j, IsWordChar)
    }

    j = StrReadWhileBackwards(text[i], j, IsSpace)

    if j < 0 {
      if i > 0 {
        i, j = i - 1, len(text[i - 1]) - 1
        if j > 0 && !IsSpace(text[i][j]) {
          found = true
          break
        }
        continue
      }
      break
    }

    found = true
    break
  }

  return i, j, found
}

func FindPrevWordStart(text Text, i, j int) (int, int, bool) {
  if j >= text.LineLen(i) {
    return i, j, false
  }

  c := text[i][j]
  if j > 0 && len(text) > 1 {
    if IsWordChar(c) && IsWordChar(text[i][j - 1]) {
      j = StrReadWhileBackwards(text[i], j, IsWordChar)
      return i, j + 1, true
    } else if IsSpecialChar(c) && IsSpecialChar(text[i][j - 1]) {
      j = StrReadWhileBackwards(text[i], j, IsSpecialChar)
      return i, j + 1, true
    }
  }

  i, j, found := FindPrevWordEnd(text, i, j)
  if found {
    c = text[i][j]

    if IsWordChar(c) {
      j = StrReadWhileBackwards(text[i], j, IsWordChar)
    } else if IsSpecialChar(c) {
      j = StrReadWhileBackwards(text[i], j, IsSpecialChar)
    }

    j = j + 1
    return i, j, true
  }

  return i, j, false
}

func FindNextWordEnd(text Text, i, j int) (int, int, bool) {
  if j >= text.LineLen(i) {
    return i, j, false
  }

  c := text[i][j]

  if j < len(text[i]) - 1 {
    if IsWordChar(c) && IsWordChar(text[i][j + 1]) {
      j = StrReadWhileForward(text[i], j, IsWordChar)
      return i, j - 1, true
    } else if IsSpecialChar(c) && IsSpecialChar(text[i][j + 1]) {
      j = StrReadWhileForward(text[i], j, IsSpecialChar)
      return i, j - 1, true
    }
  }

  i, j, found := FindNextWordStart(text, i, j)
  if found {
    c = text[i][j]

    if IsWordChar(c) {
      j = StrReadWhileForward(text[i], j, IsWordChar)
    } else if IsSpecialChar(c) {
      j = StrReadWhileForward(text[i], j, IsSpecialChar)
    }

    j = j - 1
    return i, j, true
  }

  return i, j, found
}

func FindCharForwards(text Text, ch rune, i, j int, multiline bool) (int, int, bool) {

  found := false

  if i > len(text) - 1 {
    return i, j, found
  }

  if multiline {

  multiln1:
    for i < len(text) {
      for j < len(text[i]) {
        if text[i][j] == ch {
          found = true
          break multiln1
        }
        j++
      }
      i++
      j = 0
    }

  } else {

    for j < len(text[i]) {
      if text[i][j] == ch {
        found = true
        break
      }
      j++
    }
  }

  return i, j, found
}

func FindCharBackwards(text Text, ch rune, i, j int, multiline bool) (int, int, bool) {
  found := false

  if i > len(text) - 1 {
    return i, j, found
  }

  if multiline {

  multiln2:
    for {
      for j >= 0 {
        if text[i][j] == ch {
          found = true
          break multiln2
        }
        j--
      }

      i--
      if i < 0 {
        break multiln2
      }

      j = len(text[i]) - 1
    }

  } else {

    for j >= 0 {
      if text[i][j] == ch {
        found = true
        break
      }
      j--
    }
  }

  return i, j, found
}
