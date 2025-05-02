package orchestrator

import (
	"fmt"
	"strconv"
	"strings"
)

type Node struct {
	Op    string   // Операция (+, -, *, /) или пустая строка для числа
	Value *float64 // Значение, если узел - число (лист дерева)
	Left  *Node    // Левый дочерний узел
	Right *Node    // Правый дочерний узел
}

type Parser struct {
	input string
	pos   int
	ch    byte
}

func NewParser(input string) *Parser {
	p := &Parser{input: input, pos: -1}
	p.next()
	return p
}


func (p *Parser) next() {
	p.pos++
	if p.pos < len(p.input) {
		p.ch = p.input[p.pos]
	} else {
		p.ch = 0 // Конец ввода
	}
}

func (p *Parser) skipWhitespace() {
	for p.ch != 0 && (p.ch == ' ' || p.ch == '\t' || p.ch == '\n' || p.ch == '\r') {
		p.next()
	}
}


func (p *Parser) Parse() (*Node, error) {
	if len(strings.TrimSpace(p.input)) == 0 {
		return nil, fmt.Errorf("пустое выражение")
	}
	p.pos = -1 // Сброс перед началом
	p.next()

	node, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if p.ch != 0 {
		return nil, fmt.Errorf("неожиданный символ '%c' в конце выражения", p.ch)
	}

	return node, nil
}

func (p *Parser) parseExpression() (*Node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.input) && left == nil {
		if p.ch == '+' || p.ch == '-' || p.ch == '/' || p.ch == '*' || p.ch == ')' {
			return nil, fmt.Errorf("ожидался операнд перед '%c'", p.ch)
		}
		return nil, fmt.Errorf("некорректное выражение, ожидался операнд")
	}

	for {
		p.skipWhitespace()
		if p.ch == '+' || p.ch == '-' {
			op := string(p.ch)
			p.next()
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			if right == nil {
				return nil, fmt.Errorf("ожидался операнд после '%s'", op)
			}
			left = &Node{
				Op:    op,
				Left:  left,
				Right: right,
			}
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseTerm() (*Node, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for {
		p.skipWhitespace()
		if p.ch == '*' || p.ch == '/' {
			op := string(p.ch)
			p.next()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			if right == nil {
				return nil, fmt.Errorf("ожидался операнд после '%s'", op)
			}
			left = &Node{
				Op:    op,
				Left:  left,
				Right: right,
			}
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseFactor() (*Node, error) {
	p.skipWhitespace()

	if p.ch == '-' {
		p.next()
		factor, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		if factor == nil {
			return nil, fmt.Errorf("ожидался операнд после унарного минуса")
		}
		if factor.Value != nil {
			*factor.Value = -(*factor.Value)
			return factor, nil
		} else {
			minusOne := -1.0
			return &Node{
				Op:    "*",
				Left:  &Node{Value: &minusOne},
				Right: factor,
			}, nil
		}
	}

	if p.ch == '(' {
		p.next()
		node, err := p.parseExpression() // Рекурсия для выражения в скобках
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.ch != ')' {
			return nil, fmt.Errorf("ожидалась ')', получено '%c'", p.ch)
		}
		p.next()
		return node, nil
	}

	start := p.pos
	hasDecimal := false
	for (p.ch >= '0' && p.ch <= '9') || p.ch == '.' {
		if p.ch == '.' {
			if hasDecimal {
				return nil, fmt.Errorf("некорректное число: несколько десятичных точек")
			}
			hasDecimal = true
		}
		p.next()
	}

	if start == p.pos {
		return nil, nil
	}

	numStr := p.input[start:p.pos]
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка преобразования '%s' в число: %w", numStr, err)
	}
	return &Node{Value: &val}, nil
}

func (n *Node) String() string {
	if n == nil {
		return ""
	}
	if n.Value != nil {
		v := *n.Value
		s := fmt.Sprintf("%v", v)
		if v < 0 {
			return "(" + s + ")"
		}
		return s
	}
	return fmt.Sprintf("(%s%s%s)", n.Left.String(), n.Op, n.Right.String())
}