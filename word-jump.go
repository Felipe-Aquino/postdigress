package main

import (
  "fmt"
  "strconv"
)

type Validator func(rune) bool

func StrReadWhileBackwards(text string, start int, isValid Validator) int {
  if start < 0 {
    return -1
  }

  c := rune(text[start])
  j := start

  if isValid(c) {
    for  {
      j--
      if j < 0 {
        return -1
      }
      c := rune(text[j])

      if !isValid(c) {
        return j
      }
    }
  }

  return start
}

func StrReadWhileForward(text string, start int, isValid Validator) int {
  textLen := len(text)

  if start >= textLen {
    return textLen
  }

  c := rune(text[start])
  j := start

  if isValid(c) {
    for  {
      j++
      if j > textLen - 1 {
        return j
      }
      c := rune(text[j])

      if !isValid(c) {
        return j
      }
    }
  }

  return start
}

func FindNextWordStart(text []string, i, j int) (int, int, bool) {
  found := false

  for {
    lineLen := len(text[i])

    if lineLen == 0 {
      if i < len(text) - 1 {
        i, j = i + 1, 0
        if len(text[i]) > 0 && !IsSpace(rune(text[i][j])) {
          found = true
          break
        }
        continue
      }

      break
    }

    if IsSpecialChar(rune(text[i][j])) {
      j = StrReadWhileForward(text[i], j, IsSpecialChar)
    } else if IsWordChar(rune(text[i][j])) {
      j = StrReadWhileForward(text[i], j, IsWordChar)
    }

    j = StrReadWhileForward(text[i], j, IsSpace)

    if j > lineLen - 1 {
      if i < len(text) - 1 {
        i, j = i + 1, 0
        if len(text[i]) > 0 && !IsSpace(rune(text[i][j])) {
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

func FindPrevWordEnd(text []string, i, j int) (int, int, bool) {
  found := false

  for {
    lineLen := len(text[i])

    if lineLen == 0 {
      if i > 0 {
        i, j = i - 1, len(text[i - 1]) - 1
        if j > 0 && !IsSpace(rune(text[i][j])) {
          found = true
          break
        }
        continue
      }
      break
    }

    if IsSpecialChar(rune(text[i][j])) {
      j = StrReadWhileBackwards(text[i], j, IsSpecialChar)
    } else if IsWordChar(rune(text[i][j])) {
      j = StrReadWhileBackwards(text[i], j, IsWordChar)
    }

    j = StrReadWhileBackwards(text[i], j, IsSpace)

    if j < 0 {
      if i > 0 {
        i, j = i - 1, len(text[i - 1]) - 1
        if j > 0 && !IsSpace(rune(text[i][j])) {
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

func FindPrevWordStart(text []string, i, j int) (int, int, bool) {
  c := rune(text[i][j])
  if j > 0 && len(text) > 1 {
    if IsWordChar(c) && IsWordChar(rune(text[i][j - 1])) {
      j = StrReadWhileBackwards(text[i], j, IsWordChar)
      return i, j + 1, true
    } else if IsSpecialChar(c) && IsSpecialChar(rune(text[i][j - 1])) {
      j = StrReadWhileBackwards(text[i], j, IsSpecialChar)
      return i, j + 1, true
    }
  }

  i, j, found := FindPrevWordEnd(text, i, j)
  if found {
    c = rune(text[i][j])

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

func FindNextWordEnd(text []string, i, j int) (int, int, bool) {
  c := rune(text[i][j])

  if j < len(text[i]) - 1 {
    if IsWordChar(c) && IsWordChar(rune(text[i][j + 1])) {
      j = StrReadWhileForward(text[i], j, IsWordChar)
      return i, j - 1, true
    } else if IsSpecialChar(c) && IsSpecialChar(rune(text[i][j + 1])) {
      j = StrReadWhileForward(text[i], j, IsSpecialChar)
      return i, j - 1, true
    }
  }

  i, j, found := FindNextWordStart(text, i, j)
  if found {
    c = rune(text[i][j])

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

func Replace(s string, r rune, at int) string {
  if at > len(s) - 1 {
    return s
  }

  if at == len(s) - 1 {
    return s[:at] + string(r)
  }

  return s[:at] + string(r) + s[at + 1:]
}

func word_test_main() {
  text := []string{"== word1-//-word2 word3", "", "word4"}
  i, j := 0, 0

  found := true

  mask := [3]string{}
  for l, line := range text {
    mask[l] = fmt.Sprintf("%" + strconv.Itoa(len(line)) + "s", " ")
  }

  for {
    //i, j, found = FindNextWordStart(text, i, j)
    //i, j, found = FindPrevWordEnd(text, i, j)
    //i, j, found = FindPrevWordStart(text, i, j)
    i, j, found = FindNextWordEnd(text, i, j)
    if !found {
      break
    }
    mask[i] = Replace(mask[i], '^', j)

    fmt.Printf("%d, %d\n", i, j)
    //break
  }

  for l, line := range text {
    fmt.Println(line)
    fmt.Println(mask[l])
    fmt.Print("\n")
  }
}
