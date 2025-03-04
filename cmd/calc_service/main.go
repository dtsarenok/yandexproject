package main

import (
	"encoding/json"
	"log"
	"net/http"

	"finalprogect2/internal/orchestrator"

	"github.com/gorilla/mux"

	"finalprogect2/internal/agent"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/calculate", orchestrator.AddExpression).Methods("POST")
	r.HandleFunc("/api/v1/expressions", orchestrator.GetExpressions).Methods("GET")
	r.HandleFunc("/api/v1/expressions/{id}", orchestrator.GetExpressionByID).Methods("GET")

	go agent.StartAgent()

	// Эндпоинт для получения задачи агентом
	r.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		select {
		case task := <-orchestrator.Tasks:
			json.NewEncoder(w).Encode(task)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}).Methods("GET")

	// Эндпоинт для получения результата от агента
	r.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		var result struct {
			ID     string  `json:"id"`
			Result float64 `json:"result"`
		}

		if err := json.NewDecoder(r.Body).Decode(&result); err == nil {
			orchestrator.UpdateResult(result.ID, result.Result) // Обновляем результат
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}).Methods("POST")

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
