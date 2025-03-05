package tests

import (
	"os"
	"testing"

	"github.com/Diverstt/Calculator_Yandex/internal/parser"
)

func TestParseExpression(t *testing.T) {
	exprStr := "1+2"
	ast, err := parser.ParseExpression(exprStr)
	if err != nil {
		t.Fatalf("Не удалось распарсить выражение: %v", err)
	}
	if ast == nil {
		t.Fatal("AST получен равным nil")
	}
	// Для выражения "1+2" корневой оператор должен быть "+"
	if ast.Op != "+" {
		t.Errorf("Ожидался оператор '+', получен %s", ast.Op)
	}
	// Левый и правый операнды – числовые литералы, они должны быть помечены как вычисленные
	if !ast.Left.Computed || !ast.Right.Computed {
		t.Errorf("Ожидалось, что оба операнда будут вычислены")
	}
}

func TestAssignAndFindNodeByID(t *testing.T) {
	exprStr := "1+2*3"
	ast, err := parser.ParseExpression(exprStr)
	if err != nil {
		t.Fatalf("Не удалось распарсить выражение: %v", err)
	}
	exprID := "testExpr"
	parser.AssignIDs(exprID, ast)
	// Проверяем, что корневой узел получил ID, начинающийся с "testExpr-"
	if ast.ID == "" || len(ast.ID) < len(exprID) || ast.ID[:len(exprID)] != exprID {
		t.Errorf("Узел не получил корректный ID: %s", ast.ID)
	}
	// Ищем узел по его ID
	found := parser.FindNodeByID(ast, ast.ID)
	if found == nil {
		t.Errorf("Не удалось найти узел по его ID")
	}
}

func TestGetOperationTimeAndPriority(t *testing.T) {
	// Удаляем переменные окружения для проверки значений по умолчанию.
	os.Unsetenv("TIME_ADDITION_MS")
	os.Unsetenv("TIME_SUBTRACTION_MS")
	os.Unsetenv("TIME_MULTIPLICATIONS_MS")
	os.Unsetenv("TIME_DIVISIONS_MS")

	addTime := parser.GetOperationTime("+")
	if addTime != 2000 {
		t.Errorf("Ожидалось время сложения 2000, получено %d", addTime)
	}

	subTime := parser.GetOperationTime("-")
	if subTime != 2000 {
		t.Errorf("Ожидалось время вычитания 2000, получено %d", subTime)
	}

	mulTime := parser.GetOperationTime("*")
	if mulTime != 3000 {
		t.Errorf("Ожидалось время умножения 3000, получено %d", mulTime)
	}

	divTime := parser.GetOperationTime("/")
	if divTime != 4000 {
		t.Errorf("Ожидалось время деления 4000, получено %d", divTime)
	}

	// Проверяем приоритеты операций
	if parser.GetOperationPriority("*") != 2 {
		t.Errorf("Ожидался приоритет умножения 2")
	}
	if parser.GetOperationPriority("+") != 1 {
		t.Errorf("Ожидался приоритет сложения 1")
	}
}
