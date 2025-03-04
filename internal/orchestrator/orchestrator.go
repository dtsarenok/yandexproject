package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Expression struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	Result     float64 `json:"result,omitempty"`
	Expression string  `json:"expression"`
}

type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

var (
	expressions = make(map[string]*Expression)
	Tasks       = make(chan Task, 100) // Канал для задач
	mu          sync.Mutex

	// Переменные окружения для времени выполнения операций
	TIME_ADDITION_MS        = getEnvAsInt("TIME_ADDITION_MS", 1000)
	TIME_SUBTRACTION_MS     = getEnvAsInt("TIME_SUBTRACTION_MS", 1000)
	TIME_MULTIPLICATIONS_MS = getEnvAsInt("TIME_MULTIPLICATIONS_MS", 1000)
	TIME_DIVISIONS_MS       = getEnvAsInt("TIME_DIVISIONS_MS", 1000)
)

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if valueInt, err := strconv.Atoi(valueStr); err == nil {
			return valueInt
		}
	}
	return defaultValue
}

func AddExpression(w http.ResponseWriter, r *http.Request) {
	var expr Expression
	if err := json.NewDecoder(r.Body).Decode(&expr); err != nil || expr.ID == "" || expr.Expression == "" {
		http.Error(w, "Invalid data", http.StatusUnprocessableEntity)
		return
	}

	mu.Lock()
	expressions[expr.ID] = &Expression{ID: expr.ID, Status: "pending", Expression: expr.Expression} // Сохраняем выражение
	mu.Unlock()

	go processExpression(*expressions[expr.ID]) // Обработка выражения в отдельной горутине

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": expr.ID})
}

func GetExpressions(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var result []Expression
	for _, expr := range expressions {
		result = append(result, *expr)
	}
	json.NewEncoder(w).Encode(map[string][]Expression{"expressions": result})
}

func GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/expressions/"):]

	mu.Lock()
	expr, exists := expressions[id]
	mu.Unlock()

	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]Expression{"expression": *expr})
}

func processExpression(expr Expression) {
	var arg1, arg2 float64
	var operation string

	if _, err := fmt.Sscanf(expr.Expression, "%f %s %f", &arg1, &operation, &arg2); err == nil {
		Tasks <- Task{ID: expr.ID, Arg1: arg1, Arg2: arg2, Operation: operation}
	}

	mu.Lock()
	expressions[expr.ID].Status = "in progress"
	mu.Unlock()
}

// Функция для обновления результата вычисления
func UpdateResult(taskID string, result float64) {
	mu.Lock()
	defer mu.Unlock()

	if expr, exists := expressions[taskID]; exists {
		expr.Result = result
		expr.Status = "completed"
	}
}
