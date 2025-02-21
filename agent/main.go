package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TaskResponse struct {
	Task *TaskInfo `json:"task"`
}

type TaskInfo struct {
	ID              int     `json:"id"`
	ExpressionJobID int     `json:"expression_job_id"`
	Arg1            float64 `json:"arg1"`
	Arg2            float64 `json:"arg2"`
	Operation       string  `json:"operation"`
	OperationTime   int     `json:"operation_time"`
	IsFinal         bool    `json:"is_final"`
}

type ResultRequest struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

func main() {
	computingPower := 1
	if v := os.Getenv("COMPUTING_POWER"); v != "" {
		if cp, err := strconv.Atoi(v); err == nil {
			computingPower = cp
		}
	}

	fmt.Printf("Агент стартует с %d вычислительными горутинами\n", computingPower)
	for i := 0; i < computingPower; i++ {
		go worker(i)
	}

	select {}
}

func worker(workerID int) {
	serverURL := "http://localhost:8080"
	for {
		resp, err := http.Get(serverURL + "/internal/task")
		if err != nil {
			fmt.Printf("Работник %d: ошибка получения задачи: %v\n", workerID, err)
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			time.Sleep(1 * time.Second)
			continue
		}
		var taskResp TaskResponse
		if err = json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
			fmt.Printf("Работник %d: ошибка декодирования задачи: %v\n", workerID, err)
			resp.Body.Close()
			time.Sleep(1 * time.Second)
			continue
		}
		resp.Body.Close()
		task := taskResp.Task
		result, err := compute(task.Arg1, task.Arg2, task.Operation)
		if err != nil {
			fmt.Printf("Работник %d: ошибка вычисления: %v\n", workerID, err)
			result = 0
		}
		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
		reqBody := ResultRequest{ID: task.ID, Result: result}
		data, _ := json.Marshal(reqBody)
		postResp, err := http.Post(serverURL+"/internal/task", "application/json", bytes.NewBuffer(data))
		if err != nil {
			fmt.Printf("Работник %d: ошибка отправки результата: %v\n", workerID, err)
			time.Sleep(1 * time.Second)
			continue
		}
		postResp.Body.Close()
		if postResp.StatusCode == http.StatusOK {
			fmt.Printf("%.2f %s %.2f = %.2f", task.Arg1, task.Operation, task.Arg2, result)
			if task.IsFinal {
				fmt.Printf(" (Итоговый результат: %.2f)", result)
			}
			fmt.Println()
		} else {
			fmt.Printf("Работник %d: сервер вернул код %d при отправке результата\n", workerID, postResp.StatusCode)
		}
	}
}

func compute(arg1, arg2 float64, op string) (float64, error) {
	switch op {
	case "+":
		return arg1 + arg2, nil
	case "-":
		return arg1 - arg2, nil
	case "*":
		return arg1 * arg2, nil
	case "/":
		if arg2 == 0 {
			return 0, fmt.Errorf("деление на ноль")
		}
		return arg1 / arg2, nil
	default:
		return 0, fmt.Errorf("неизвестная операция: %s", op)
	}
}
