package agent

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

var (
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

func StartAgent() {
	// Получаем количество горутин из переменной среды
	computingPower, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))

	var wg sync.WaitGroup
	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				task := getTask()
				if task != nil {
					result := performOperation(task)
					sendResult(task.ID, result)
				}
				time.Sleep(1 * time.Second) // Задержка между запросами задач
			}
		}()
	}

	wg.Wait()
}

func getTask() *Task {
	resp, err := http.Get("http://localhost/internal/task")
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	var task Task
	json.NewDecoder(resp.Body).Decode(&task)
	return &task
}

func performOperation(task *Task) float64 {
	time.Sleep(time.Duration(getOperationTime(task.Operation)+rand.Intn(100)) * time.Millisecond)

	switch task.Operation {
	case "add":
		return task.Arg1 + task.Arg2
	case "subtract":
		return task.Arg1 - task.Arg2
	case "multiply":
		return task.Arg1 * task.Arg2
	case "divide":
		return task.Arg1 / task.Arg2
	default:
		return 0.0
	}
}

func getOperationTime(operation string) int {
	switch operation {
	case "add":
		return TIME_ADDITION_MS
	case "subtract":
		return TIME_SUBTRACTION_MS
	case "multiply":
		return TIME_MULTIPLICATIONS_MS
	case "divide":
		return TIME_DIVISIONS_MS
	default:
		return 0
	}
}

func sendResult(taskID string, result float64) {
	data := map[string]interface{}{
		"id":     taskID,
		"result": result,
	}
	body, _ := json.Marshal(data)
	http.Post("http://localhost/internal/task", "application/json", bytes.NewBuffer(body))
}
