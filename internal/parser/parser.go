package parser

import (
	"fmt"
	"go/ast"
	goParser "go/parser"
	"log"
	"os"
	"strconv"
	"strings"
)

type Node struct {
	ID        string
	Op        string
	Value     float64
	Left      *Node
	Right     *Node
	Parent    *Node
	Computed  bool
	Scheduled bool
}

func (n *Node) IsReady() bool {
	if n.Left == nil || n.Right == nil {
		return false
	}
	return n.Left.Computed && n.Right.Computed
}

func ParseExpression(expression string) (*Node, error) {
	expr, err := goParser.ParseExpr(expression)
	if err != nil {
		return nil, err
	}
	return buildAST(expr)
}

func buildAST(expr ast.Expr) (*Node, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		value, err := strconv.ParseFloat(strings.Trim(e.Value, "\""), 64)
		if err != nil {
			return nil, fmt.Errorf("неверное число: %s", e.Value)
		}
		return &Node{
			Value:    value,
			Computed: true,
		}, nil

	case *ast.BinaryExpr:
		left, err := buildAST(e.X)
		if err != nil {
			return nil, err
		}
		right, err := buildAST(e.Y)
		if err != nil {
			return nil, err
		}
		node := &Node{
			Op:       e.Op.String(),
			Left:     left,
			Right:    right,
			Computed: false,
		}
		left.Parent = node
		right.Parent = node
		return node, nil

	case *ast.ParenExpr:
		return buildAST(e.X)

	default:
		return nil, fmt.Errorf("неподдерживаемый тип выражения: %T", e)
	}
}

func AssignIDs(exprID string, node *Node) {
	var counter int
	assignIDsRecursive(exprID, node, &counter, nil)
}

func assignIDsRecursive(exprID string, node *Node, counter *int, parent *Node) {
	if node == nil {
		return
	}
	*counter++
	node.ID = fmt.Sprintf("%s-%d", exprID, *counter)
	node.Parent = parent

	if node.Left != nil {
		assignIDsRecursive(exprID, node.Left, counter, node)
	}
	if node.Right != nil {
		assignIDsRecursive(exprID, node.Right, counter, node)
	}
}

func FindNodeByID(node *Node, id string) *Node {
	if node == nil {
		return nil
	}
	if node.ID == id {
		return node
	}
	if found := FindNodeByID(node.Left, id); found != nil {
		return found
	}
	return FindNodeByID(node.Right, id)
}

func GetOperationTime(op string) int {
	var envVar string
	switch op {
	case "+":
		envVar = "TIME_ADDITION_MS"
	case "-":
		envVar = "TIME_SUBTRACTION_MS"
	case "*":
		envVar = "TIME_MULTIPLICATIONS_MS"
	case "/":
		envVar = "TIME_DIVISIONS_MS"
	default:
		return 1000
	}

	valStr := os.Getenv(envVar)
	if valStr == "" {
		switch op {
		case "+", "-":
			return 2000
		case "*":
			return 3000
		case "/":
			return 4000
		default:
			return 1000
		}
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Ошибка преобразования %s: %v", envVar, err)
		switch op {
		case "+", "-":
			return 2000
		case "*":
			return 3000
		case "/":
			return 4000
		default:
			return 1000
		}
	}
	return val
}

func GetOperationPriority(op string) int {
	switch op {
	case "*", "/":
		return 2
	case "+", "-":
		return 1
	default:
		return 0
	}
}
