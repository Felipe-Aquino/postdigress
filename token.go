package main

import (
  "fmt"
  "strings"
)

type TokenType int

const (
  IDENT  TokenType = iota
  NUMBER
  STRING

  TYPE
  KEYWORD
  COMMENT

  OTHER
  READ_END
)

  /*
    "boolean",  "char", "character", "varchar",   "date",
    "precision", "integer",   "int", "numeric",   "decimal",   "real",
    "smallint",  "timestamp",
  */
var types [13]uint32 = [13]uint32{
    1710517951, 2335876880, 2823553821, 4163743794, 3564297305,
    2214008087, 3218261061, 2515107422, 1761125480, 520654156, 3604983901,
    2174562837, 2994984227,
  }

  /*
  hashes of:

    "add",    "all",     "alter",   "from",   "any",      "replace",
    "asc",    "between", "by",      "case",   "check",    "column",
    "create", "left",    "default", "delete", "desc",     "where",
    "and",    "drop",    "exec",    "exists", "foreign",  "distinct",
    "full",   "group",   "having",  "in",     "index",    "inner",
    "insert", "into",    "is",      "join",   "key",      "database",
    "like",   "limit",   "not",     "null",   "or",       "order",
    "outer",  "primary", "view",    "as",     "right",    "rownum",
    "select", "set",     "table",   "top",    "truncate", "union",
    "unique", "update",  "values",  "constraint", "procedure", 
  */
var keywords [59]uint32 = [59]uint32{
   993596020,  321211332,  583954283,  2513272949, 740945997,  2704835779,
   577951402,  2767062895, 1412156564, 2602907825, 3010534471, 4126724647,
   649812317,  306900080,  2470140894, 1740784714, 2198011936, 90969176,
   254395046,  2846199180, 3742059100, 1002329533, 1389469171, 719363213,
   4286165820, 1605967500, 3672956252, 1094220446, 151693739,  3638454823,
   3332609576, 798074659,  1312329493, 3374496889, 1746258028, 2707807672,
   199741238,  853203252,  699505802,  1996966820, 1563699588, 1932267671,
   2718680436, 1512988633, 3685020920, 1579491469, 2028154341, 1029694425,
   297952813, 3324446467, 1251777503, 2802900028, 1111263159, 3688814324,
   1035877158, 672109684, 877087803,  3089293914, 2800173518,
  }

func (tt TokenType) String() string {
  switch tt {
  case IDENT:
    return "IDENT"
  case NUMBER:
    return "NUMBER"
  case STRING:
    return "STRING"
  case TYPE:
    return "TYPE"
  case KEYWORD:
    return "KEYWORD"
  case COMMENT:
    return "COMMENT"
  case OTHER:
    return "OTHER"
  case READ_END:
    return "READ_END"
  }
  return "??";
}

func GetIdentType(name string) TokenType {
  h := Hash(strings.ToLower(name))
  for _, t := range types {
    if h == t {
      return TYPE
    }
  }

  for _, k := range keywords {
    if h == k {
      return KEYWORD
    }
  }

  return IDENT
}

type Token struct {
  ttype TokenType
  line, col int

  start, size int
}

func NewToken(ttype TokenType, line, col, start, size int) Token {
  return Token{ ttype, line, col, start, size }
}

func (t Token) Value(text *string) string {
  return (*text)[t.start: t.start + t.size]
}

func (t Token) Is(ttype TokenType) bool {
  return t.ttype == ttype
}

func (t Token) Print(text *string) {
  fmt.Printf(
    "{ type: %s, value: %s, line: %d, col: %d }\n",
    t.ttype.String(),
    t.Value(text),
    t.line,
    t.col)
}

type Lock struct {
  pos, line, col int 
  active bool

};

func NewLock(pos, line, col int, active bool) Lock {
  return Lock{pos, line, col, active}
}

type Tokenizer struct {
  pos int
  line, col int

  input string

  current Token

  lock Lock
}

func NewTokenizer() *Tokenizer {
  return &Tokenizer{pos: 0, line: 1, col: 1, input: "", lock: Lock{} }
}

func (tn *Tokenizer) SetInput(input string) {
  tn.pos = 0
  tn.line = 0
  tn.col = 0
  tn.input = input
}

func (tn *Tokenizer) MakeToken(ttype TokenType, size int) Token {
  return NewToken(ttype, tn.line, tn.col, tn.pos, size)
}

func (tn *Tokenizer) Lock() {
  tn.lock = NewLock(tn.pos, tn.line, tn.col, true)
}

func (tn *Tokenizer) Commit() {
  tn.lock.active = false
}

func (tn *Tokenizer) Rollback() {
  tn.lock.active = false
  tn.pos = tn.lock.pos
  tn.line = tn.lock.line
  tn.col = tn.lock.col
}

func (tn *Tokenizer) LockPosDiff() int {
  if tn.lock.active {
    return tn.pos - tn.lock.pos
  }
  return 0
}

func (tn *Tokenizer) LockGetData() string {
  return tn.input[tn.lock.pos: tn.pos]
}

func (tn *Tokenizer) IsEnd() bool {
  return tn.pos >= len(tn.input)
}

func (tn *Tokenizer) EatSpaces() {
  for tn.pos < len(tn.input) && IsSpace(rune(tn.input[tn.pos])) {
    if tn.input[tn.pos] == '\n' {
      tn.line++
      tn.col = 0
    }

    tn.col++
    tn.pos++
  }
}

func (tn *Tokenizer) ReadNumber() {
  c := rune(tn.input[tn.pos])

  if IsDigit(c) {
    for {
      tn.pos++
      if tn.IsEnd() {
        break
      }

      c = rune(tn.input[tn.pos])

      if !IsDigit(c) {
        break
      }
    }
  }

  if (c == '.') {
    for {
      tn.pos++
      if tn.IsEnd() {
        break
      }

      c = rune(tn.input[tn.pos])

      if !IsDigit(c) {
        break
      }
    }
  }
}

func (tn *Tokenizer) ReadString(delim rune) {
  c := rune(tn.input[tn.pos])

  if c == delim {
    c := rune(tn.input[tn.pos + 1])

    for !tn.IsEnd() && c != delim {
      if c == '\n' {
        tn.col = 0
        tn.line++
      }

      tn.col++
      tn.pos++

      if tn.pos < len(tn.input) {
        c = rune(tn.input[tn.pos])
      }
    }

    if c == delim {
      tn.pos += 1
    }
  }
}

func (tn *Tokenizer) ReadIdent() {
  c := rune(tn.input[tn.pos])

  if c == '_' || IsAlpha(c) {
    for !tn.IsEnd() && (IsAlnum(c) || c == '_') {
      tn.pos++
      c = rune(tn.input[tn.pos])
    }
  }
}

func (tn *Tokenizer) ReadInlineComment() {
  if tn.input[tn.pos: tn.pos + 2] == "--" {
    c := tn.input[tn.pos]
    tn.pos++

    for !tn.IsEnd() && (c != '\n') {
      c = tn.input[tn.pos]
      tn.pos++
    }
  }
}

func (tn *Tokenizer) ReadMultilineComment() {
  match := tn.input[tn.pos: tn.pos + 2]

  if match == "/*" {
    match = tn.input[tn.pos + 1: tn.pos + 3]

    for !tn.IsEnd() && match != "*/" {
      if match[0] == '\n' {
        tn.col = 0
        tn.line++
      }

      tn.col++
      tn.pos++
      if tn.pos < len(tn.input) - 1 {
        match = tn.input[tn.pos: tn.pos + 2]
      }
    }

    if match == "*/" {
      tn.pos += 2
    }
  }
}

func (tn *Tokenizer) NextToken() Token {
  tn.EatSpaces()

  token := tn.MakeToken(OTHER, 1)

  if tn.pos >= len(tn.input) {
    token.ttype = READ_END
    return token
  }

  c := rune(tn.input[tn.pos])

  next_c := rune(0)
  if tn.pos + 1 < len(tn.input) {
    next_c = rune(tn.input[tn.pos + 1])
  }

  if c == '-' && next_c == '-' {
    tn.Lock()
    tn.ReadInlineComment()
    token.ttype = COMMENT
    token.size = tn.LockPosDiff()
    tn.Commit()

  } else if c == '/' && next_c == '*' {
    tn.Lock()
    tn.ReadMultilineComment()
    token.ttype = COMMENT
    token.size = tn.LockPosDiff()
    tn.Commit()

  } else if IsDigit(c) || (c == '.' && IsDigit(next_c)) {
    tn.Lock()
    tn.ReadNumber()
    token.ttype = NUMBER
    token.size = tn.LockPosDiff()
    tn.Commit()

  } else if c == '"' || c == '\'' {
    tn.Lock()
    tn.ReadString(c)
    token.ttype = STRING
    token.size = tn.LockPosDiff()
    tn.Commit()

  } else if c == '_' || IsAlpha(c) {
    tn.Lock()
    tn.ReadIdent()
    token.ttype = GetIdentType(tn.LockGetData())
    token.size = tn.LockPosDiff()
    tn.Commit()
  } else {
    tn.pos += token.size; 
  }

  //fmt.Printf("pos: %d\n", tn.pos)

  tn.current = token

  tn.col += token.size
  return token
}

func _main() {
  tokenizer := NewTokenizer()

  text := "-- just a comment\n select * from abc;"
  tokenizer.SetInput(text)

  for !tokenizer.IsEnd() {
    token := tokenizer.NextToken()
    token.Print(&text)
  }
}
