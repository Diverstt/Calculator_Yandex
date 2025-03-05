// internal/models/task.go
package models

// Task описывает отдельную арифметическую операцию, которую необходимо вычислить.
type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"` // время выполнения операции в мс
	Priority      int     `json:"priority"`       // приоритет вычисления (чем выше значение, тем приоритетнее)
}
