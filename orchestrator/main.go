package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var mu sync.Mutex
var jobs = make(map[int]*ExpressionJob)
var tasksQueue []*Task
var tasksInProgress = make(map[int]*Task)
var nextJobID = 1
var nextTaskID = 1
var timeAddition int
var timeSubtraction int
var timeMultiplication int
var timeDivision int

type ExpressionJob struct {
	ID         int      `json:"id"`
	Expression string   `json:"expression"` // можно использовать для отладки
	Status     string   `json:"status"`
	Result     *float64 `json:"result"`
	AST        *Node    `json:"-"`
	Steps      []string `json:"steps"`
}

type Node struct {
	Op        string
	Value     *float64
	Left      *Node
	Right     *Node
	Parent    *Node
	Scheduled bool
}

type Task struct {
	ID              int
	ExpressionJobID int
	Node            *Node
	Operation       string
	Arg1            float64
	Arg2            float64
	OperationTime   int
}

type Parser struct {
	input string
	pos   int
	ch    byte
}

func NewParser(input string) *Parser {
	p := &Parser{input: input, pos: -1}
	p.next()
	return p
}

func (p *Parser) next() {
	p.pos++
	if p.pos < len(p.input) {
		p.ch = p.input[p.pos]
	} else {
		p.ch = 0
	}
}

func (p *Parser) skipWhitespace() {
	for p.ch != 0 && (p.ch == ' ' || p.ch == '\t' || p.ch == '\n' || p.ch == '\r') {
		p.next()
	}
}

func (p *Parser) parseExpression() (*Node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.ch == '+' || p.ch == '-' {
			op := string(p.ch)
			p.next()
			p.skipWhitespace()
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			node := &Node{
				Op:    op,
				Left:  left,
				Right: right,
			}
			if left != nil {
				left.Parent = node
			}
			if right != nil {
				right.Parent = node
			}
			left = node
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseTerm() (*Node, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.ch == '*' || p.ch == '/' {
			op := string(p.ch)
			p.next()
			p.skipWhitespace()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			node := &Node{
				Op:    op,
				Left:  left,
				Right: right,
			}
			if left != nil {
				left.Parent = node
			}
			if right != nil {
				right.Parent = node
			}
			left = node
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseFactor() (*Node, error) {
	p.skipWhitespace()
	if p.ch == '(' {
		p.next()
		node, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.ch != ')' {
			return nil, fmt.Errorf("ожидалась ')'")
		}
		p.next()
		return node, nil
	}
	start := p.pos
	for (p.ch >= '0' && p.ch <= '9') || p.ch == '.' {
		p.next()
	}
	if start == p.pos {
		return nil, fmt.Errorf("ожидалось число")
	}
	numStr := p.input[start:p.pos]
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, err
	}
	return &Node{Value: &val}, nil
}

func scheduleReadyTasks(node *Node, jobID int) {
	if node == nil {
		return
	}
	if node.Op == "*" || node.Op == "/" {
		planNodeTask(node, jobID)
	}
	scheduleReadyTasks(node.Left, jobID)
	scheduleReadyTasks(node.Right, jobID)
	if node.Op == "+" || node.Op == "-" {
		planNodeTask(node, jobID)
	}
}

func planNodeTask(node *Node, jobID int) {
	if node.Value != nil || node.Scheduled || node.Op == "" {
		return
	}
	leftReady := node.Left != nil && node.Left.Value != nil
	rightReady := node.Right != nil && node.Right.Value != nil
	if leftReady && rightReady {
		node.Scheduled = true
		task := &Task{
			ID:              nextTaskID,
			ExpressionJobID: jobID,
			Node:            node,
			Operation:       node.Op,
			Arg1:            getArgValue(node.Left),
			Arg2:            getArgValue(node.Right),
			OperationTime:   getOperationTime(node.Op),
		}
		nextTaskID++
		tasksQueue = append(tasksQueue, task)
	}
}

func getArgValue(node *Node) float64 {
	if node == nil {
		return 0
	}
	if node.Value != nil {
		return *node.Value
	}
	return 0
}

func getOperationTime(op string) int {
	switch op {
	case "+":
		return timeAddition
	case "-":
		return timeSubtraction
	case "*":
		return timeMultiplication
	case "/":
		return timeDivision
	default:
		return 1000
	}
}

func internalTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		mu.Lock()
		if len(tasksQueue) == 0 {
			mu.Unlock()
			http.Error(w, "Нет задач", http.StatusNotFound)
			return
		}
		task := tasksQueue[0]
		tasksQueue = tasksQueue[1:]
		tasksInProgress[task.ID] = task
		mu.Unlock()
		response := map[string]interface{}{
			"task": map[string]interface{}{
				"id":                task.ID,
				"expression_job_id": task.ExpressionJobID,
				"arg1":              task.Arg1,
				"arg2":              task.Arg2,
				"operation":         task.Operation,
				"operation_time":    task.OperationTime,
				"is_final":          task.Node.Parent == nil,
			},
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	if r.Method == "POST" {
		var reqBody struct {
			ID     int     `json:"id"`
			Result float64 `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Невалидные данные", http.StatusUnprocessableEntity)
			return
		}
		mu.Lock()
		task, ok := tasksInProgress[reqBody.ID]
		if !ok {
			mu.Unlock()
			http.Error(w, "Задача не найдена", http.StatusNotFound)
			return
		}
		delete(tasksInProgress, reqBody.ID)
		node := task.Node
		node.Value = &reqBody.Result
		if node.Parent == nil {
			job, exists := jobs[task.ExpressionJobID]
			if exists {
				job.Result = node.Value
				job.Status = "done"
				step := fmt.Sprintf("%.2f %s %.2f = %.2f", task.Arg1, task.Operation, task.Arg2, reqBody.Result)
				if job.Steps == nil {
					job.Steps = make([]string, 0)
				}
				job.Steps = append(job.Steps, step)
			}
		} else {
			job, exists := jobs[task.ExpressionJobID]
			if exists {
				step := fmt.Sprintf("%.2f %s %.2f = %.2f", task.Arg1, task.Operation, task.Arg2, reqBody.Result)
				if job.Steps == nil {
					job.Steps = make([]string, 0)
				}
				job.Steps = append(job.Steps, step)
			}
			scheduleReadyTasks(node.Parent, task.ExpressionJobID)
		}
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var reqData struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	expr := strings.TrimSpace(reqData.Expression)

	if expr == "" {
		http.Error(w, "Пустое выражение", http.StatusBadRequest)
		return
	}

	parser := NewParser(expr)
	ast, err := parser.parseExpression()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка парсинга: %v", err), http.StatusBadRequest)
		return
	}

	job := &ExpressionJob{
		ID:         nextJobID,
		Expression: expr,
		Status:     "in_progress",
		AST:        ast,
	}
	nextJobID++
	jobs[job.ID] = job
	scheduleReadyTasks(ast, job.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         job.ID,
		"expression": job.Expression,
		"status":     job.Status,
	})
}

func expressionsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
		if idStr == "" {
			var jobList []ExpressionJob
			for _, job := range jobs {
				jobList = append(jobList, *job)
			}
			json.NewEncoder(w).Encode(jobList)
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Неверный id", http.StatusBadRequest)
			return
		}
		mu.Lock()
		job, exists := jobs[id]
		mu.Unlock()
		if !exists {
			http.Error(w, "Выражение не найдено", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	case "POST":
		var job ExpressionJob
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			http.Error(w, "Невалидные данные", http.StatusBadRequest)
			return
		}
		job.ID = nextJobID
		nextJobID++
		jobs[job.ID] = &job
		json.NewEncoder(w).Encode(job)
	}
}

func initOperationTimes() {
	if v := os.Getenv("TIME_ADDITION_MS"); v != "" {
		if t, err := strconv.Atoi(v); err == nil {
			timeAddition = t
		} else {
			timeAddition = 1000
		}
	} else {
		timeAddition = 1000
	}
	if v := os.Getenv("TIME_SUBTRACTION_MS"); v != "" {
		if t, err := strconv.Atoi(v); err == nil {
			timeSubtraction = t
		} else {
			timeSubtraction = 1000
		}
	} else {
		timeSubtraction = 1000
	}
	if v := os.Getenv("TIME_MULTIPLICATIONS_MS"); v != "" {
		if t, err := strconv.Atoi(v); err == nil {
			timeMultiplication = t
		} else {
			timeMultiplication = 1000
		}
	} else {
		timeMultiplication = 1000
	}
	if v := os.Getenv("TIME_DIVISIONS_MS"); v != "" {
		if t, err := strconv.Atoi(v); err == nil {
			timeDivision = t
		} else {
			timeDivision = 1000
		}
	} else {
		timeDivision = 1000
	}
}

func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

func main() {
	initOperationTimes()

	// Оборачиваем обработчики в enableCORS
	http.HandleFunc("/api/v1/calculate", enableCORS(calculateHandler))
	http.HandleFunc("/api/v1/expressions", enableCORS(expressionsRouter))
	http.HandleFunc("/api/v1/expressions/", enableCORS(expressionsRouter))
	http.HandleFunc("/internal/task", internalTaskHandler)

	fmt.Println("Оркестратор слушает на порту :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Ошибка старта сервера:", err)
	}
}
