package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fp(f float64) *float64 {
	return &f
}

func TestParser(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"Простое сложение", "2 + 3", 5, false},
		{"Сложное выражение", "2 * (3 + 4)", 14, false},
		{"Некорректное выражение", "2 + + 3", 0, true},
		{"Скобки", "(2 + 3) * 4", 20, false},
		{"Пустое выражение", "", 0, true},
		{"Неправильные скобки", "(2 + 3", 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.parseExpression()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && node == nil {
				t.Error("parseExpression() вернул nil, ожидалось валидное дерево")
			}
		})
	}
}

func TestScheduleReadyTasks(t *testing.T) {
	tests := []struct {
		name         string
		expr         string
		wantTask     bool
		expectedOp   string
		expectedArg1 float64
		expectedArg2 float64
	}{
		{"Простое выражение", "2 + 3", true, "+", 2, 3},
		{"Сложное выражение", "2 * (3 + 4)", true, "+", 3, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu.Lock()
			defer mu.Unlock()
			tasksQueue = []*Task{}
			parser := NewParser(tt.expr)
			ast, err := parser.parseExpression()
			if err != nil {
				t.Fatalf("Ошибка парсинга: %v", err)
			}
			scheduleReadyTasks(ast, 1)
			if tt.wantTask && len(tasksQueue) == 0 {
				t.Error("scheduleReadyTasks() не создал задачи, хотя должен был")
				return
			}
			if tt.wantTask && len(tasksQueue) > 0 {
				task := tasksQueue[0]
				if task.Operation != tt.expectedOp {
					t.Errorf("Ожидалась операция %s, получена %s", tt.expectedOp, task.Operation)
				}
				if task.Arg1 != tt.expectedArg1 {
					t.Errorf("Ожидался аргумент1 %f, получен %f", tt.expectedArg1, task.Arg1)
				}
				if task.Arg2 != tt.expectedArg2 {
					t.Errorf("Ожидался аргумент2 %f, получен %f", tt.expectedArg2, task.Arg2)
				}
			}
		})
	}
}

func TestScheduleReadyTasksPriority(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantOps []string
	}{
		{"Приоритет умножения", "2 + 2 * 2", []string{"*"}},
		{"Приоритет деления", "2 + 6 / 2", []string{"/"}},
		{"Сложное выражение", "2 * 3 + 4 * 5", []string{"*", "*"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mu.Lock()
			tasksQueue = []*Task{}
			mu.Unlock()
			parser := NewParser(tt.expr)
			ast, err := parser.parseExpression()
			if err != nil {
				t.Fatalf("Ошибка парсинга: %v", err)
			}
			scheduleReadyTasks(ast, 1)
			if len(tasksQueue) != len(tt.wantOps) {
				t.Errorf("Ожидалось %d задач, получено %d", len(tt.wantOps), len(tasksQueue))
				return
			}
			for i, op := range tt.wantOps {
				if tasksQueue[i].Operation != op {
					t.Errorf("Задача %d: ожидалась операция %s, получена %s", i, op, tasksQueue[i].Operation)
				}
			}
		})
	}
}

func TestInternalTaskGet(t *testing.T) {
	mu.Lock()
	tasksQueue = []*Task{}
	tasksInProgress = make(map[int]*Task)
	dummyNode := &Node{
		Op:    "+",
		Value: nil,
		Left:  &Node{Value: fp(2)},
		Right: &Node{Value: fp(3)},
	}
	dummyTask := &Task{
		ID:              999,
		ExpressionJobID: 42,
		Node:            dummyNode,
		Operation:       "+",
		Arg1:            2,
		Arg2:            3,
		OperationTime:   100,
	}
	tasksQueue = append(tasksQueue, dummyTask)
	mu.Unlock()
	req := httptest.NewRequest("GET", "/internal/task", nil)
	rr := httptest.NewRecorder()
	internalTaskHandler(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус 200, получен %d", res.StatusCode)
	}
	var respBody map[string]map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}
	taskMap := respBody["task"]
	if id, ok := taskMap["id"].(float64); !ok || int(id) != 999 {
		t.Errorf("Ожидался id задачи 999, получено %v", taskMap["id"])
	}
	if ejid, ok := taskMap["expression_job_id"].(float64); !ok || int(ejid) != 42 {
		t.Errorf("Ожидался expression_job_id 42, получено %v", taskMap["expression_job_id"])
	}
}

func TestInternalTaskPost(t *testing.T) {
	mu.Lock()
	tasksInProgress = make(map[int]*Task)
	dummyNode := &Node{
		Op:    "+",
		Value: nil,
		Left:  &Node{Value: fp(2)},
		Right: &Node{Value: fp(3)},
	}
	dummyTask := &Task{
		ID:              1001,
		ExpressionJobID: 50,
		Node:            dummyNode,
		Operation:       "+",
		Arg1:            2,
		Arg2:            3,
		OperationTime:   50,
	}
	tasksInProgress[dummyTask.ID] = dummyTask
	dummyJob := &ExpressionJob{
		ID:         50,
		Expression: "2+3",
		Status:     "in_progress",
		AST:        dummyNode,
	}
	jobs[dummyJob.ID] = dummyJob
	mu.Unlock()
	payload := `{"id":1001, "result":5.0}`
	req := httptest.NewRequest("POST", "/internal/task", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	internalTaskHandler(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус 200, получен %d", res.StatusCode)
	}
	mu.Lock()
	if _, exists := tasksInProgress[1001]; exists {
		t.Error("Задача не удалена из tasksInProgress")
	}
	updatedJob, jobExists := jobs[50]
	if !jobExists {
		t.Error("Выражение не найдено")
	} else if updatedJob.Status != "done" || updatedJob.Result == nil || *updatedJob.Result != 5.0 {
		t.Errorf("Неверное обновление выражения: %+v", updatedJob)
	}
	mu.Unlock()
}
