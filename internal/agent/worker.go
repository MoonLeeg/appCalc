package agent

import (
	pb "calculator/internal/grpc/calculator" // Обновленный импорт gRPC кода
	"context"
	"fmt"
	"log"
	"time"
)

func Worker(workerID int, grpcClient pb.CalculatorAgentServiceClient) {
	log.Printf("Воркер %d запущен.", workerID)
	ctx := context.Background() // Основной контекст для gRPC вызовов
	agentID := fmt.Sprintf("agent-%d", workerID)

	for {
		log.Printf("Воркер %d: Запрос задачи...", workerID)
		var task *pb.Task
		var err error
		var retryAfter time.Duration = 1 * time.Second // Задержка по умолчанию

		getTaskReq := &pb.GetTaskRequest{AgentId: agentID}
		getTaskResp, err := grpcClient.GetTask(ctx, getTaskReq)

		if err != nil {
			log.Printf("Воркер %d: Ошибка gRPC при получении задачи: %v. Повтор через %v...", workerID, err, retryAfter)
			time.Sleep(retryAfter)
			continue
		}

		switch taskInfo := getTaskResp.TaskInfo.(type) {
		case *pb.GetTaskResponse_Task:
			task = taskInfo.Task
			log.Printf("Воркер %d: Получена задача ID %d: %f %s %f (время: %dms)",
				workerID, task.Id, task.Arg1, task.Operation, task.Arg2, task.OperationTimeMs)
		case *pb.GetTaskResponse_NoTask:
			if taskInfo.NoTask != nil && taskInfo.NoTask.RetryAfterSeconds > 0 {
				retryAfter = time.Duration(taskInfo.NoTask.RetryAfterSeconds) * time.Second
			}
			log.Printf("Воркер %d: Нет доступных задач. Повтор через %v...", workerID, retryAfter)
			time.Sleep(retryAfter)
			continue // Переходим к следующей итерации цикла
		default:
			log.Printf("Воркер %d: Получен неизвестный ответ от GetTask. Повтор через %v...", workerID, retryAfter)
			time.Sleep(retryAfter)
			continue
		}

		startTime := time.Now()
		result, computeErr := compute(task.Arg1, task.Arg2, task.Operation)
		computationDuration := time.Since(startTime)

		if task.OperationTimeMs > 0 {
			requiredDuration := time.Duration(task.OperationTimeMs) * time.Millisecond
			if computationDuration < requiredDuration {
				time.Sleep(requiredDuration - computationDuration)
			}
		}

		submitReq := &pb.SubmitResultRequest{
			TaskId:  task.Id,
			AgentId: agentID,
		}
		if computeErr != nil {
			log.Printf("Воркер %d: Ошибка вычисления задачи ID %d: %v", workerID, task.Id, computeErr)
			submitReq.ResultStatus = &pb.SubmitResultRequest_Error{
				Error: &pb.TaskError{Message: computeErr.Error()},
			}
		} else {
			log.Printf("Воркер %d: Завершено вычисление задачи ID %d. Результат: %f", workerID, task.Id, result)
			submitReq.ResultStatus = &pb.SubmitResultRequest_Result{Result: result}
		}

		_, err = grpcClient.SubmitResult(ctx, submitReq)
		if err != nil {
			log.Printf("Воркер %d: Ошибка gRPC при отправке результата задачи ID %d: %v. Задача может быть переназначена.", workerID, task.Id, err)
			time.Sleep(retryAfter) // Небольшая пауза перед запросом новой задачи
		} else {
			log.Printf("Воркер %d: Результат задачи ID %d успешно отправлен.", workerID, task.Id)
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