package orchestrator

import (
	"container/heap"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Diverstt/Calculator_Yandex/internal/models"
	"github.com/Diverstt/Calculator_Yandex/internal/parser"
)

type TaskPriorityQueue []*models.Task

func (pq TaskPriorityQueue) Len() int {
	return len(pq)
}

func (pq TaskPriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority
}

func (pq TaskPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *TaskPriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*models.Task))
}
func (pq *TaskPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]

	return item
}

type Server struct {
	Router      *http.ServeMux
	Expressions map[string]*models.Expression
	ASTs        map[string]*parser.Node
	TaskQueue   TaskPriorityQueue
	QueueMutex  sync.Mutex
	Mutex       sync.Mutex
}

func NewServer() *Server {
	s := &Server{
		Router:      http.NewServeMux(),
		Expressions: make(map[string]*models.Expression),
		ASTs:        make(map[string]*parser.Node),
		TaskQueue:   make(TaskPriorityQueue, 0),
	}
	heap.Init(&s.TaskQueue)
	s.Router.HandleFunc("/api/v1/calculate", s.handleCalculate)
	s.Router.HandleFunc("/api/v1/expressions", s.handleExpressions)
	s.Router.HandleFunc("/api/v1/expressions/", s.handleExpressionByID)
	s.Router.HandleFunc("/internal/task", s.handleTask)
	s.Router.HandleFunc("/internal/task/result", s.handleTaskResult)

	return s
}

func (s *Server) handleCalculate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusUnprocessableEntity)
		return
	}

	exprID := time.Now().Format("20060102150405")
	expr := &models.Expression{ID: exprID, Status: "pending"}

	ast, err := parser.ParseExpression(input.Expression)
	if err != nil {
		http.Error(w, "Неверное арифметическое выражение", http.StatusUnprocessableEntity)
		return
	}

	parser.AssignIDs(exprID, ast)

	s.Mutex.Lock()
	s.Expressions[exprID] = expr
	s.ASTs[exprID] = ast
	s.Mutex.Unlock()

	log.Printf("Выражение %s принято: %s", exprID, input.Expression)

	s.scheduleReadyTasks(exprID, ast)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": exprID})
}

func (s *Server) handleExpressions(w http.ResponseWriter, r *http.Request) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	expressions := make([]models.Expression, 0, len(s.Expressions))
	for _, expr := range s.Expressions {
		expressions = append(expressions, *expr)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions})
}

func (s *Server) handleExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	expr, ok := s.Expressions[id]
	if !ok {
		http.Error(w, "Выражение не найдено", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
	s.QueueMutex.Lock()
	if s.TaskQueue.Len() == 0 {
		s.QueueMutex.Unlock()
		http.Error(w, "Нет доступных задач", http.StatusNotFound)
		return
	}

	task := heap.Pop(&s.TaskQueue).(*models.Task)
	s.QueueMutex.Unlock()
	log.Printf("Задача %s отправлена агенту: %+v", task.ID, task)
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (s *Server) handleTaskResult(w http.ResponseWriter, r *http.Request) {
	var res models.Result
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusUnprocessableEntity)
		return
	}

	// Если пришла ошибка вычисления (например, деление на ноль)
	if res.Error != "" {
		s.Mutex.Lock()
		// Обновляем статус выражения на "error" для узла, связанного с данной задачей
		for exprID, root := range s.ASTs {
			node := parser.FindNodeByID(root, res.ID)
			if node != nil {
				if expr, ok := s.Expressions[exprID]; ok {
					expr.Status = "error"
				}
				break
			}
		}
		s.Mutex.Unlock()
		http.Error(w, res.Error, http.StatusUnprocessableEntity)
		return
	}

	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	found := false
	for exprID, root := range s.ASTs {
		node := parser.FindNodeByID(root, res.ID)
		if node != nil {
			node.Value = res.Result
			node.Computed = true
			log.Printf("Обновлен узел %s: результат %f", node.ID, res.Result)
			if node.Parent != nil && node.Parent.IsReady() && !node.Parent.Scheduled {
				task := &models.Task{
					ID:            node.Parent.ID,
					Arg1:          node.Parent.Left.Value,
					Arg2:          node.Parent.Right.Value,
					Operation:     node.Parent.Op,
					OperationTime: parser.GetOperationTime(node.Parent.Op),
					Priority:      parser.GetOperationPriority(node.Parent.Op),
				}
				node.Parent.Scheduled = true
				s.QueueMutex.Lock()
				heap.Push(&s.TaskQueue, task)
				s.QueueMutex.Unlock()
				log.Printf("Запланирована задача для узла %s родителя", node.Parent.ID)
			} else if node.Parent == nil {
				if expr, ok := s.Expressions[exprID]; ok {
					expr.Result = &res.Result
					expr.Status = "completed"
					log.Printf("Выражение %s полностью вычислено: %f", exprID, res.Result)
				}
			}
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "результат записан"})
}

func (s *Server) scheduleReadyTasks(exprID string, node *parser.Node) {
	if node == nil {
		return
	}
	if !node.Computed && node.IsReady() && !node.Scheduled {
		task := &models.Task{
			ID:            node.ID,
			Arg1:          node.Left.Value,
			Arg2:          node.Right.Value,
			Operation:     node.Op,
			OperationTime: parser.GetOperationTime(node.Op),
			Priority:      parser.GetOperationPriority(node.Op),
		}
		node.Scheduled = true
		s.QueueMutex.Lock()
		heap.Push(&s.TaskQueue, task)
		s.QueueMutex.Unlock()
		log.Printf("Запланирована задача для узла %s: %f %s %f, приоритет %d", node.ID, node.Left.Value, node.Op, node.Right.Value, task.Priority)
	}
	s.scheduleReadyTasks(exprID, node.Left)
	s.scheduleReadyTasks(exprID, node.Right)
}
