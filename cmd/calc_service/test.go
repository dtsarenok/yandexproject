package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finalprogect2/internal/orchestrator"

	"github.com/gorilla/mux"
)

func TestAddExpression(t *testing.T) {
	reqBody := map[string]string{
		"expression": "2+2",
		"id":         "test-id",
	}

	reqBodyJSON, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/calculate", orchestrator.AddExpression).Methods("POST")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp["id"] != "test-id" {
		t.Errorf("Expected id %s, got %s", "test-id", resp["id"])
	}
}

func TestGetExpressions(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/expressions", orchestrator.GetExpressions).Methods("GET")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string][]orchestrator.Expression
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp["expressions"]) == 0 {
		t.Errorf("Expected non-empty list of expressions")
	}
}

func TestGetExpressionByID(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/expressions/test-id", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/expressions/{id}", orchestrator.GetExpressionByID).Methods("GET")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]orchestrator.Expression
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp["expression"].ID != "test-id" {
		t.Errorf("Expected id %s, got %s", "test-id", resp["expression"].ID)
	}
}

func TestGetTask(t *testing.T) {
	req, err := http.NewRequest("GET", "/internal/task", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		select {
		case task := <-orchestrator.Tasks:
			json.NewEncoder(w).Encode(task)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}).Methods("GET")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSendResult(t *testing.T) {
	reqBody := map[string]interface{}{
		"id":     "test-id",
		"result": 4.0,
	}

	reqBodyJSON, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/internal/task", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := mux.NewRouter()
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
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
