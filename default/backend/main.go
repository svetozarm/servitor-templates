package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

func dataFile() string {
	return filepath.Join("data", "todos.json")
}

func loadTodos() []Todo {
	b, err := os.ReadFile(dataFile())
	if err != nil {
		return []Todo{}
	}
	var todos []Todo
	json.Unmarshal(b, &todos)
	return todos
}

func saveTodos(todos []Todo) {
	os.MkdirAll("data", 0755)
	b, _ := json.Marshal(todos)
	os.WriteFile(dataFile(), b, 0644)
}

func writeResponse(status int, body []byte) {
	fmt.Printf("HTTP/1.1 %d %s\r\n", status, http.StatusText(status))
	fmt.Printf("Content-Type: application/json\r\n")
	fmt.Printf("Content-Length: %d\r\n", len(body))
	fmt.Printf("\r\n")
	os.Stdout.Write(body)
}

func main() {
	req, err := http.ReadRequest(bufio.NewReader(os.Stdin))
	if err != nil {
		writeResponse(400, []byte(`{"error":"bad request"}`))
		return
	}

	path := req.URL.Path
	method := req.Method

	switch {
	case path == "/api/todos" && method == "GET":
		b, _ := json.Marshal(loadTodos())
		writeResponse(200, b)

	case path == "/api/todos" && method == "POST":
		var input struct{ Text string `json:"text"` }
		json.NewDecoder(req.Body).Decode(&input)
		todos := loadTodos()
		id := fmt.Sprintf("%d", len(todos)+1)
		todos = append(todos, Todo{ID: id, Text: input.Text})
		saveTodos(todos)
		writeResponse(201, []byte(`{"ok":true}`))

	case strings.HasPrefix(path, "/api/todos/") && method == "PATCH":
		id := strings.TrimPrefix(path, "/api/todos/")
		todos := loadTodos()
		for i := range todos {
			if todos[i].ID == id {
				todos[i].Done = !todos[i].Done
				break
			}
		}
		saveTodos(todos)
		writeResponse(200, []byte(`{"ok":true}`))

	case strings.HasPrefix(path, "/api/todos/") && method == "DELETE":
		id := strings.TrimPrefix(path, "/api/todos/")
		todos := loadTodos()
		filtered := todos[:0]
		for _, t := range todos {
			if t.ID != id {
				filtered = append(filtered, t)
			}
		}
		saveTodos(filtered)
		writeResponse(200, []byte(`{"ok":true}`))

	default:
		writeResponse(404, []byte(`{"error":"not found"}`))
	}
}
