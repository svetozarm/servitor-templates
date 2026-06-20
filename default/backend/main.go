package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Todo struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

const dataDir = "data/todos"

func genID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func todoPath(id string) string {
	return filepath.Join(dataDir, id+".json")
}

func loadTodo(id string) (Todo, bool) {
	b, err := os.ReadFile(todoPath(id))
	if err != nil {
		return Todo{}, false
	}
	var t Todo
	json.Unmarshal(b, &t)
	return t, true
}

func loadAllTodos() []Todo {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return []Todo{}
	}
	todos := make([]Todo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dataDir, e.Name()))
		if err != nil {
			continue
		}
		var t Todo
		json.Unmarshal(b, &t)
		todos = append(todos, t)
	}
	return todos
}

func saveTodo(t Todo) {
	os.MkdirAll(dataDir, 0755)
	b, _ := json.Marshal(t)
	os.WriteFile(todoPath(t.ID), b, 0644)
}

func writeResponse(status int, body []byte) {
	fmt.Printf("HTTP/1.1 %d %s\r\n", status, http.StatusText(status))
	fmt.Printf("Content-Type: application/json\r\n")
	fmt.Printf("Content-Length: %d\r\n", len(body))
	fmt.Printf("\r\n")
	os.Stdout.Write(body)
}

func writeNoContent() {
	fmt.Printf("HTTP/1.1 204 No Content\r\n")
	fmt.Printf("\r\n")
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
		b, _ := json.Marshal(loadAllTodos())
		writeResponse(200, b)

	case path == "/api/todos" && method == "POST":
		var input struct {
			Text string `json:"text"`
		}
		json.NewDecoder(req.Body).Decode(&input)
		if input.Text == "" {
			writeResponse(422, []byte(`{"error":"text is required"}`))
			return
		}
		t := Todo{ID: genID(), Text: input.Text, Done: false, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
		saveTodo(t)
		b, _ := json.Marshal(t)
		writeResponse(201, b)

	case strings.HasPrefix(path, "/api/todos/") && method == "GET":
		id := strings.TrimPrefix(path, "/api/todos/")
		t, ok := loadTodo(id)
		if !ok {
			writeResponse(404, []byte(`{"error":"not found"}`))
			return
		}
		b, _ := json.Marshal(t)
		writeResponse(200, b)

	case strings.HasPrefix(path, "/api/todos/") && method == "PUT":
		id := strings.TrimPrefix(path, "/api/todos/")
		t, ok := loadTodo(id)
		if !ok {
			writeResponse(404, []byte(`{"error":"not found"}`))
			return
		}
		var input struct {
			Text string `json:"text"`
			Done *bool  `json:"done"`
		}
		json.NewDecoder(req.Body).Decode(&input)
		if input.Text != "" {
			t.Text = input.Text
		}
		if input.Done != nil {
			t.Done = *input.Done
		}
		saveTodo(t)
		b, _ := json.Marshal(t)
		writeResponse(200, b)

	case strings.HasPrefix(path, "/api/todos/") && method == "DELETE":
		id := strings.TrimPrefix(path, "/api/todos/")
		if _, ok := loadTodo(id); !ok {
			writeResponse(404, []byte(`{"error":"not found"}`))
			return
		}
		os.Remove(todoPath(id))
		writeNoContent()

	default:
		writeResponse(404, []byte(`{"error":"not found"}`))
	}
}
