package main

type TokenType string

type Token struct {
	TokenType TokenType
	Literal   string
	Line      int
	Col       int
}

const (
	//EO
	T_EOF = "EOF"
	T_EOL = "EOL"
	//keywords
	T_TABLE       = "table"
	T_END         = "end"
	T_SET         = "set"
	T_LEFT_PAREN  = "LeftParan"
	T_RIGHT_PAREN = "RightParan"
	T_COMMA       = "Comma"
	//other
	T_ATTR    = "Attr"
	T_IDEN    = "Iden"
	T_STRING  = "String"
	T_RAW     = "Raw"
	T_NUM     = "Number"
	T_ILLEGAL = "Illegal"
	T_ERROR   = "Error"
)

func createToken(tokenType TokenType, value string, line int, col int) *Token {
	return &Token{
		TokenType: tokenType,
		Literal:   value,
		Line:      line,
		Col:       col,
	}
}

func getToken(literal string, line int, col int) *Token {
	var tokenType TokenType = T_IDEN

	switch literal {
	case T_TABLE:
		tokenType = T_TABLE
		break
	case T_END:
		tokenType = T_END
	case T_SET:
		tokenType = T_SET
	default:
		break
	}

	return createToken(tokenType, literal, line, col)
}

type Tokenizer struct {
	input string
	pos   int
	line  int
	col   int
	ch    byte
}

func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input,
		0,
		1,
		1,
		'\x00',
	}
}

func (t *Tokenizer) NextToken() *Token {

	ch := t.readChar()

	// avoiding pos and col increase when end of file get reached
	if ch == '\x00' {
		return createToken(T_EOF, T_EOF, t.line, t.col)
	}

	//save a copy of where token starts
	startCol := t.col

	t.nextChar()

	//skip whitespace
	if isWhitespace(ch) {
		return t.NextToken()
	}

	switch ch {
	case '\n':
		tok := createToken(T_EOL, T_EOL, t.line, t.col)
		t.jumpLine()
		return tok
	case '@':
		//checking if after @ is a whitespace if not create token and skip @ char
		ch = t.readChar()
		t.nextChar()
		if isWhitespace(ch) {
			break
		} else {
			literal := t.readIden()
			return createToken(T_ATTR, literal, t.line, startCol)
		}
	case '(':
		return createToken(T_LEFT_PAREN, "(", t.line, startCol)
	case ')':
		return createToken(T_RIGHT_PAREN, ")", t.line, startCol)
	case ',':
		return createToken(T_COMMA, ",", t.line, startCol)
	case '"':
		literal := t.readString('"')
		return createToken(T_STRING, literal, t.line, startCol)
	case '`':
		literal := t.readString('`')
		return createToken(T_RAW, literal, t.line, startCol)
	}

	if isLetter(ch) {
		literal := t.readIden()
		return getToken(literal, t.line, startCol)
	}

	if isNumber(ch) {
		number := t.readNum()
		return createToken(T_NUM, number, t.line, startCol)
	}

	//default value
	return createToken(T_ILLEGAL, string(ch), t.line, startCol)
}

func (t *Tokenizer) PeekToken() *Token {
	line := t.line
	col := t.col
	pos := t.pos
	tok := t.NextToken()
	t.line = line
	t.col = col
	t.pos = pos
	return tok
}

func (t *Tokenizer) readChar() byte {
	if t.isEOF() {
		return '\x00'
	}

	ch := t.input[t.pos]

	if isEOL(ch) {
		return '\n'
	}

	t.ch = ch
	return ch
}

func (t *Tokenizer) skipWhitespace() {
	t.readChar()
	if isWhitespace(t.ch) {
		t.nextChar()
		t.skipWhitespace()
	}
}

func (t *Tokenizer) nextChar() {
	t.pos++
	t.col++
}

func (t *Tokenizer) readIden() string {
	iden := []rune{rune(t.ch)}
	for {
		ch := t.readChar()

		if !isLetter(ch) && !isNumber(ch) {
			break
		}

		iden = append(iden, rune(ch))

		t.nextChar()
	}

	return string(iden)
}

func (t *Tokenizer) readNum() string {
	num := []rune{rune(t.ch)}

	for {
		ch := t.readChar()
		if !isNumber(ch) {
			break
		}
		num = append(num, rune(ch))
		t.nextChar()
	}

	return string(num)
}

func (t *Tokenizer) readString(delim byte) string {
	str := []rune{}

	for {
		ch := t.readChar()
		t.nextChar()

		if ch == '\x00' {
			//unterminated string
			break
		}

		if ch == delim {
			break
		}

		if isEOL(ch) {
			t.jumpLine()
		}

		str = append(str, rune(ch))
	}

	return string(str)
}

func (t *Tokenizer) jumpLine() {
	t.col = 1
	t.line++
}

// end of file
func (t *Tokenizer) isEOF() bool {
	return t.pos >= len(t.input)
}

// end of line
func isEOL(char byte) bool {
	return char == '\n' || char == '\r'
}

func isNumber(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

// for debuging
func (t *Tokenizer) Reset() {
	t.pos = 0
	t.line = 1
	t.col = 1
}
