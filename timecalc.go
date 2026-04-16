package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type expression struct {
	Head *term       `parser:"@@"`
	Tail []*exprTail `parser:"@@*"`
}

type exprTail struct {
	Op    string `parser:"@(\"+\" | \"-\")"`
	Right *term  `parser:"@@"`
}

type term struct {
	Head *unary      `parser:"@@"`
	Tail []*termTail `parser:"@@*"`
}

type termTail struct {
	Op    string       `parser:"@(\"*\" | \"/\")"`
	Right *scalarUnary `parser:"@@"`
}

type unary struct {
	Signs []string `parser:"@(\"+\" | \"-\")*"`
	Value *primary `parser:"@@"`
}

type scalarUnary struct {
	Signs []string       `parser:"@(\"+\" | \"-\")*"`
	Value *scalarPrimary `parser:"@@"`
}

type primary struct {
	Group   *expression `parser:"  \"(\" @@ \")\""`
	Literal *literal    `parser:"| @@"`
}

type scalarPrimary struct {
	Number string `parser:"@Number"`
}

type literal struct {
	Time   *string `parser:"  @Time"`
	Number *string `parser:"| @Number"`
}

var (
	expressionLexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Time", Pattern: `\d+:\d+`},
		{Name: "Number", Pattern: `(?:\d+(?:\.\d*)?|\.\d+)`},
		{Name: "Plus", Pattern: `\+`},
		{Name: "Minus", Pattern: `-`},
		{Name: "Multiply", Pattern: `\*`},
		{Name: "Divide", Pattern: `/`},
		{Name: "LParen", Pattern: `\(`},
		{Name: "RParen", Pattern: `\)`},
		{Name: "Whitespace", Pattern: `\s+`},
	})

	expressionParser = participle.MustBuild[expression](
		participle.Lexer(expressionLexer),
		participle.Elide("Whitespace"),
	)
)

func parseTimeInput(s string) (int, error) {
	if strings.Contains(s, ":") {
		return parseTimeToMinutes(s)
	}

	floatVal, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("無効な時間形式: %s", s)
	}
	mins := int(floatVal * 60)
	return mins, nil
}

func parseTimeToMinutes(s string) (int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("無効な形式: %s", s)
	}
	hours, err1 := strconv.Atoi(parts[0])
	minutes, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || minutes < 0 || minutes >= 60 {
		return 0, fmt.Errorf("無効な時刻値: %s", s)
	}
	return hours*60 + minutes, nil
}

func formatMinutesWithDecimal(total int) string {
	sign := ""
	minutesAbs := total
	if total < 0 {
		sign = "-"
		minutesAbs = -total
	}
	h := minutesAbs / 60
	m := minutesAbs % 60
	decimal := float64(minutesAbs) / 60.0
	return fmt.Sprintf("%s%d:%02d (%.3f)", sign, h, m, decimal)
}

func normalizeTokens(tokens []string) []string {
	var normalized []string
	for _, t := range tokens {
		switch strings.ToLower(t) {
		case "p":
			normalized = append(normalized, "+")
		case "m":
			normalized = append(normalized, "-")
		default:
			normalized = append(normalized, t)
		}
	}
	return normalized
}

func (expr *expression) eval() (int, error) {
	result, err := expr.Head.eval()
	if err != nil {
		return 0, err
	}

	for _, tail := range expr.Tail {
		right, err := tail.Right.eval()
		if err != nil {
			return 0, err
		}
		switch tail.Op {
		case "+":
			result += right
		case "-":
			result -= right
		}
	}
	return result, nil
}

func (termNode *term) eval() (int, error) {
	result, err := termNode.Head.eval()
	if err != nil {
		return 0, err
	}

	for _, tail := range termNode.Tail {
		right, err := tail.Right.eval()
		if err != nil {
			return 0, err
		}
		switch tail.Op {
		case "*":
			result = int(float64(result) * right)
		case "/":
			if right == 0 {
				return 0, fmt.Errorf("0では割れません")
			}
			result = int(float64(result) / right)
		}
	}
	return result, nil
}

func (unaryNode *unary) eval() (int, error) {
	value, err := unaryNode.Value.eval()
	if err != nil {
		return 0, err
	}

	for i := len(unaryNode.Signs) - 1; i >= 0; i-- {
		if unaryNode.Signs[i] == "-" {
			value = -value
		}
	}
	return value, nil
}

func (unaryNode *scalarUnary) eval() (float64, error) {
	value, err := unaryNode.Value.eval()
	if err != nil {
		return 0, err
	}

	for i := len(unaryNode.Signs) - 1; i >= 0; i-- {
		if unaryNode.Signs[i] == "-" {
			value = -value
		}
	}
	return value, nil
}

func (primaryNode *primary) eval() (int, error) {
	if primaryNode.Group != nil {
		return primaryNode.Group.eval()
	}
	return primaryNode.Literal.eval()
}

func (scalar *scalarPrimary) eval() (float64, error) {
	value, err := strconv.ParseFloat(scalar.Number, 64)
	if err != nil {
		return 0, fmt.Errorf("無効な数値: %s", scalar.Number)
	}
	return value, nil
}

func (lit *literal) eval() (int, error) {
	if lit.Time != nil {
		return parseTimeToMinutes(*lit.Time)
	}
	if lit.Number != nil {
		return parseTimeInput(*lit.Number)
	}
	return 0, fmt.Errorf("式が不正です")
}

func evaluateExpression(tokens []string) (int, error) {
	expression := strings.Join(tokens, " ")
	ast, err := expressionParser.ParseString("", expression)
	if err != nil {
		return 0, err
	}
	return ast.eval()
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("使い方: timecalc 式")
		return
	}

	tokens := normalizeTokens(args)
	resultMinutes, err := evaluateExpression(tokens)
	if err != nil {
		fmt.Println("エラー:", err)
		return
	}

	fmt.Println(formatMinutesWithDecimal(resultMinutes))
}
