package main

import "hash/fnv"

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

type Rect struct {
  x, y, width, height int
}

func (r Rect) Unpack() (int, int, int, int) {
  return r.x, r.y, r.width, r.height
}

func (r Rect) YMax() int {
  return r.y + r.height
}

func (r Rect) XMax() int {
  return r.x + r.width
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

func DigitNumber(n int) int {
  if n < 10 {
    return 1
  } else if n < 100 {
    return 2
  }
  return 3
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

func RuneSliceClamp(s []rune, min, max int) []rune {
  size := len(s)
  if size > min {
    if size > max {
      return s[min:max]
    }
    return s[min:]
  }
  return []rune{}
}

// Split string in N pieces
func StrSplitInN(s string, n int) ([]string, bool) {
	pieces := []string{}

  if n < 1 {
    return pieces, true
  }

	nPieces := len(s) / n

	start := 0
	for i := 0; i < nPieces; i++ {
		pieces = append(pieces, s[start:(i + 1)*n])
		start = (i + 1) * n
	}

	if start < len(s) {
		pieces = append(pieces, s[start:])
	}

  return pieces, false
}

func U8Charsize(ch int) int {
  if ch <= 0x7F {
    return 1
  } else if ch <= 0xE0 {
    return 2
  } else if ch <= 0xF0 {
    return 3
  }
  return 4
}

func SSMap(vs [][]string, f func([]string) string) []string {
  vsm := make([]string, len(vs))
  for i, v := range vs {
    vsm[i] = f(v)
  }
  return vsm
}
