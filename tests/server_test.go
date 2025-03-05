package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Diverstt/Calculator_Yandex/internal/models"
	"github.com/Diverstt/Calculator_Yandex/internal/orchestrator"
)

func TestCalculateEndpoint(t *testing.T) {
	server := orchestrator.NewServer()
	ts := httptest.NewServer(server.Router)
	defer ts.Close()

	expression := "1+2"
	payload := map[string]string{"expression": expression}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/calculate", "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("Ошибка при вызове /api/v1/calculate: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Ожидался статус 201 Created, получен %d", resp.StatusCode)
	}
	var res map[string]string
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Не удалось декодировать ответ: %v", err)
	}
	exprID, ok := res["id"]
	if !ok || exprID == "" {
		t.Fatalf("Не получен ID выражения")
	}

	time.Sleep(100 * time.Millisecond)

	resp, err = http.Get(ts.URL + "/internal/task")
	if err != nil {
		t.Fatalf("Ошибка при запросе задачи: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус OK для /internal/task, получен %d", resp.StatusCode)
	}
	var taskRes map[string]models.Task
	err = json.NewDecoder(resp.Body).Decode(&taskRes)
	if err != nil {
		t.Fatalf("Не удалось декодировать ответ задачи: %v", err)
	}
	task, ok := taskRes["task"]
	if !ok {
		t.Fatalf("В ответе не найдена задача")
	}
	if task.Operation != "+" {
		t.Errorf("Ожидалась операция '+', получена %s", task.Operation)
	}

	resultPayload := models.Result{
		ID:     task.ID,
		Result: 3,
	}
	resultData, _ := json.Marshal(resultPayload)
	resp, err = http.Post(ts.URL+"/internal/task/result", "application/json", bytes.NewBuffer(resultData))
	if err != nil {
		t.Fatalf("Ошибка при отправке результата задачи: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус OK от /internal/task/result, получен %d: %s", resp.StatusCode, string(body))
	}

	resp, err = http.Get(ts.URL + "/api/v1/expressions/" + exprID)
	if err != nil {
		t.Fatalf("Ошибка при запросе выражения по ID: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус OK для получения выражения, получен %d", resp.StatusCode)
	}
	var exprRes map[string]models.Expression
	err = json.NewDecoder(resp.Body).Decode(&exprRes)
	if err != nil {
		t.Fatalf("Не удалось декодировать ответ выражения: %v", err)
	}
	exprData, ok := exprRes["expression"]
	if !ok {
		t.Fatalf("Выражение не найдено в ответе")
	}
	if exprData.Status != "completed" {
		t.Errorf("Ожидался статус 'completed', получен '%s'", exprData.Status)
	}
	if exprData.Result == nil || *exprData.Result != 3 {
		t.Errorf("Ожидался результат 3, получено %v", exprData.Result)
	}
}

func TestExpressionsEndpoint(t *testing.T) {
	server := orchestrator.NewServer()
	ts := httptest.NewServer(server.Router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/expressions")
	if err != nil {
		t.Fatalf("Ошибка при запросе списка выражений: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус OK, получен %d", resp.StatusCode)
	}
	var listRes map[string][]models.Expression
	err = json.NewDecoder(resp.Body).Decode(&listRes)
	if err != nil {
		t.Fatalf("Не удалось декодировать список выражений: %v", err)
	}
	expressions := listRes["expressions"]
	if len(expressions) != 0 {
		t.Errorf("Ожидалось 0 выражений, получено %d", len(expressions))
	}

	expression := "4-2"
	payload := map[string]string{"expression": expression}
	data, _ := json.Marshal(payload)
	resp, err = http.Post(ts.URL+"/api/v1/calculate", "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("Ошибка при вызове /api/v1/calculate: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Ожидался статус 201 Created, получен %d", resp.StatusCode)
	}
	var res map[string]string
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Не удалось декодировать ответ: %v", err)
	}
	exprID := res["id"]

	time.Sleep(100 * time.Millisecond)

	resp, err = http.Get(ts.URL + "/api/v1/expressions")
	if err != nil {
		t.Fatalf("Ошибка при запросе списка выражений: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус OK, получен %d", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&listRes)
	if err != nil {
		t.Fatalf("Не удалось декодировать список выражений: %v", err)
	}
	expressions = listRes["expressions"]
	if len(expressions) != 1 {
		t.Errorf("Ожидалось 1 выражение, получено %d", len(expressions))
	}

	found := false
	for _, expr := range expressions {
		if strings.EqualFold(expr.ID, exprID) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Выражение с ID %s не найдено в списке", exprID)
	}
}

func TestDivisionByZeroEndpoint(t *testing.T) {
	server := orchestrator.NewServer()
	ts := httptest.NewServer(server.Router)
	defer ts.Close()

	// Создаём выражение с делением на ноль, например "1/0"
	expression := "1/0"
	payload := map[string]string{"expression": expression}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/calculate", "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("Ошибка при вызове /api/v1/calculate: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Ожидался статус 201 Created, получен %d", resp.StatusCode)
	}
	var res map[string]string
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Не удалось декодировать ответ: %v", err)
	}
	exprID, ok := res["id"]
	if !ok || exprID == "" {
		t.Fatalf("Не получен ID выражения")
	}

	time.Sleep(100 * time.Millisecond)

	// Получаем задачу, которая должна быть операцией деления
	resp, err = http.Get(ts.URL + "/internal/task")
	if err != nil {
		t.Fatalf("Ошибка при запросе задачи: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус OK для /internal/task, получен %d", resp.StatusCode)
	}
	var taskRes map[string]models.Task
	err = json.NewDecoder(resp.Body).Decode(&taskRes)
	if err != nil {
		t.Fatalf("Не удалось декодировать ответ задачи: %v", err)
	}
	task, ok := taskRes["task"]
	if !ok {
		t.Fatalf("В ответе не найдена задача")
	}
	if task.Operation != "/" {
		t.Fatalf("Ожидалась операция '/', получена %s", task.Operation)
	}

	// Симулируем отправку результата с ошибкой деления на ноль
	resultPayload := models.Result{
		ID:     task.ID,
		Result: 0,
		Error:  "деление на ноль",
	}
	resultData, _ := json.Marshal(resultPayload)
	resp, err = http.Post(ts.URL+"/internal/task/result", "application/json", bytes.NewBuffer(resultData))
	if err != nil {
		t.Fatalf("Ошибка при отправке результата задачи: %v", err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус 422 Unprocessable Entity, получен %d: %s", resp.StatusCode, string(body))
	}

	// Проверяем, что статус выражения обновлен на "error"
	resp, err = http.Get(ts.URL + "/api/v1/expressions/" + exprID)
	if err != nil {
		t.Fatalf("Ошибка при запросе выражения по ID: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус OK для получения выражения, получен %d", resp.StatusCode)
	}
	var exprRes map[string]models.Expression
	err = json.NewDecoder(resp.Body).Decode(&exprRes)
	if err != nil {
		t.Fatalf("Не удалось декодировать ответ выражения: %v", err)
	}
	exprData, ok := exprRes["expression"]
	if !ok {
		t.Fatalf("Выражение не найдено в ответе")
	}
	if exprData.Status != "error" {
		t.Errorf("Ожидался статус 'error', получен '%s'", exprData.Status)
	}
}
