package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Diverstt/Calculator_Yandex/internal/models"
)

var orchestratorURL string

func init() {
	orchestratorURL = os.Getenv("ORCHESTRATOR_URL")
	if orchestratorURL == "" {
		orchestratorURL = "http://orchestrator:8080"
	}
}

func StartWorkers(count int) {
	for i := 0; i < count; i++ {
		go worker(i + 1)
	}
}

func worker(id int) {
	log.Printf("Агент #%d запущен", id)
	for {
		task, err := fetchTask()
		if err != nil {
			log.Printf("Агент #%d: ошибка при получении задачи: %v", id, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if task == nil {
			time.Sleep(2 * time.Second)
			continue
		}
		log.Printf("Агент #%d получил задачу: %+v", id, task)

		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

		result, err := compute(task)
		if err != nil {
			log.Printf("Агент #%d: ошибка при вычислении задачи %s: %v", id, task.ID, err)
			sendResult(task.ID, 0, err.Error())
		} else {
			log.Printf("Агент #%d вычислил результат задачи %s: %f", id, task.ID, result)
			sendResult(task.ID, result, "")
		}
	}
}

func fetchTask() (*models.Task, error) {
	url := fmt.Sprintf("%s/internal/task", orchestratorURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var response struct {
		Task models.Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response.Task, nil
}

func compute(task *models.Task) (float64, error) {
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2, nil
	case "-":
		return task.Arg1 - task.Arg2, nil
	case "*":
		return task.Arg1 * task.Arg2, nil
	case "/":
		if task.Arg2 == 0 {
			return 0, fmt.Errorf("деление на ноль")
		}
		return task.Arg1 / task.Arg2, nil
	default:
		return 0, fmt.Errorf("неподдерживаемая операция: %s", task.Operation)
	}
}

func sendResult(taskID string, result float64, errMsg string) error {
	res := models.Result{
		ID:     taskID,
		Result: result,
		Error:  errMsg,
	}
	data, _ := json.Marshal(res)
	url := fmt.Sprintf("%s/internal/task/result", orchestratorURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
