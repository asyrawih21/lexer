package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"unicode"
)

type Token int

const (
	EOF = iota
	ILLEGAL
	IDENT
	INT

	// Infix ops
	ADD // +
	SUB // -
	MUL // *
	DIV // /
)

var tokens = []string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",
	IDENT:   "IDENT",
	INT:     "INT",
	ADD:     "+",
	SUB:     "-",
	MUL:     "*",
	DIV:     "/",
}

type Position struct {
	line   int
	column int
}

type Lexer struct {
	pos    Position
	reader *bufio.Reader
}

func (l *Lexer) Lex() (Token, string) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return EOF, ""
			}
			log.Fatal(err)
		}
		l.pos.column++

		switch r {
		case '\n':
			l.resetPosition()
		case '+':
			return ADD, "+"
		case '-':
			return SUB, "-"
		case '*':
			return MUL, "*"
		case '/':
			return DIV, "/"
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsDigit(r) {
				l.backup()
				lit := l.lexInt()
				return INT, lit
			} else {
				return ILLEGAL, string(r)
			}
		}
	}
}

func (l *Lexer) resetPosition() {
	l.pos.line++
	l.pos.column = 0
}

func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		log.Fatal(err)
	}
	l.pos.column--
}

func (l *Lexer) lexInt() string {
	var lit string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return lit
			}
			log.Fatal(err)
		}
		l.pos.column++
		if unicode.IsDigit(r) {
			lit = lit + string(r)
		} else {
			l.backup()
			return lit
		}
	}
}

func NewLexer(reader *bufio.Reader) *Lexer {
	return &Lexer{
		pos:    Position{line: 1, column: 0},
		reader: bufio.NewReader(reader),
	}
}

type Node interface {
	Pos() Position
	String() string
}

type Expression interface {
	Node
	exprNode()
}

type BinaryExpression struct {
	Left     Expression
	Op       Token
	Right    Expression
	Position Position
}

func (be *BinaryExpression) Pos() Position {
	return be.Position
}

func (be *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", be.Left.String(), tokens[be.Op], be.Right.String())
}

func (be *BinaryExpression) exprNode() {}

type IntegerLiteral struct {
	Value    int
	Position Position
}

// exprNode implements Expression.
func (*IntegerLiteral) exprNode() {
	panic("unimplemented")
}

func (il *IntegerLiteral) Pos() Position {
	return il.Position
}

func (il *IntegerLiteral) String() string {
	return strconv.Itoa(il.Value)
}

func parseExpression(l *Lexer) Expression {
	return parseAddSubExpr(l)
}

func parseAddSubExpr(l *Lexer) Expression {
	left := parseMulDivExpr(l)

	for {
		tok, _ := l.Lex()
		if tok != ADD && tok != SUB {
			l.backup()
			return left
		}

		right := parseMulDivExpr(l)
		left = &BinaryExpression{Left: left, Op: tok, Right: right, Position: left.Pos()}
	}
}

func parseMulDivExpr(l *Lexer) Expression {
	left := parsePrimaryExpr(l)

	for {
		tok, _ := l.Lex()
		if tok != MUL && tok != DIV {
			l.backup()
			return left
		}

		right := parsePrimaryExpr(l)
		left = &BinaryExpression{Left: left, Op: tok, Right: right, Position: left.Pos()}
	}
}

func parsePrimaryExpr(l *Lexer) Expression {
	tok, lit := l.Lex()

	if tok == INT {
		value, _ := strconv.Atoi(lit)
		return &IntegerLiteral{Value: value, Position: Position{line: l.pos.line, column: l.pos.column}}
	}

	log.Fatalf("Unexpected token: %s", tokens[tok])
	return nil // unreachable
}

func evaluateExpression(expr Expression) (int, error) {
	switch e := expr.(type) {
	case *BinaryExpression:
		left, err := evaluateExpression(e.Left)
		if err != nil {
			return 0, err
		}

		right, err := evaluateExpression(e.Right)
		if err != nil {
			return 0, err
		}

		switch e.Op {
		case ADD:
			return left + right, nil
		case SUB:
			return left - right, nil
		case MUL:
			return left * right, nil
		case DIV:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		default:
			return 0, fmt.Errorf("unknown operator")
		}

	case *IntegerLiteral:
		return e.Value, nil

	default:
		return 0, fmt.Errorf("unknown expression type")
	}
}

func main() {
	r := bufio.NewReader(os.Stdin)
	l := NewLexer(r)

	expr := parseExpression(l)
	fmt.Println(expr)
	result, err := evaluateExpression(expr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
