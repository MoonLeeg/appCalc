package orchestrator

import (
	"calculator/internal/database"
	pb "calculator/internal/grpc/calculator" // Обновленный импорт gRPC кода
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedCalculatorAgentServiceServer // Встраивание для обратной совместимости
	dbStore                                      *database.Store
	opTimes                                      *OperationTimes // Нужны для заполнения operation_time_ms в задаче
	scheduler                                    *Scheduler      // Добавляем планировщик для обработки завершения
}

func NewCalculatorGRPCServer(db *database.Store, opTimes *OperationTimes, scheduler *Scheduler) *grpcServer {
	return &grpcServer{
		dbStore:   db,
		opTimes:   opTimes,
		scheduler: scheduler, // Сохраняем планировщик
	}
}

func (s *grpcServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	log.Printf("gRPC: Получен запрос GetTask от агента ID: %s", req.AgentId)

	task, err := s.dbStore.GetAndLeasePendingTask()
	if err != nil {
		log.Printf("gRPC: Ошибка получения задачи из БД: %v", err)
		return nil, status.Errorf(codes.Internal, "ошибка БД при получении задачи: %v", err)
	}

	if task == nil {
		log.Println("gRPC: Нет доступных задач для агента")
		return &pb.GetTaskResponse{
			TaskInfo: &pb.GetTaskResponse_NoTask{
				NoTask: &pb.NoTaskAvailable{
					RetryAfterSeconds: 5, // Говорим агенту попробовать через 5 секунд
				},
			},
		}, nil
	}

	log.Printf("gRPC: Отправка задачи ID %d агенту %s", task.ID, req.AgentId)
	return &pb.GetTaskResponse{
		TaskInfo: &pb.GetTaskResponse_Task{
			Task: &pb.Task{
				Id:              task.ID,
				Arg1:            task.Arg1,
				Arg2:            task.Arg2,
				Operation:       task.Operation,
				OperationTimeMs: s.getOperationTimeMs(task.Operation), // Получаем время для операции
			},
		},
	}, nil
}

func (s *grpcServer) SubmitResult(ctx context.Context, req *pb.SubmitResultRequest) (*pb.SubmitResultResponse, error) {
	log.Printf("gRPC: Получен результат SubmitResult для задачи ID %d от агента ID: %s", req.TaskId, req.AgentId)
	var taskErr error

	switch result := req.ResultStatus.(type) {
	case *pb.SubmitResultRequest_Result:
		taskErr = s.dbStore.CompleteTask(req.TaskId, result.Result)
		if taskErr == nil {
			log.Printf("gRPC: Задача ID %d успешно завершена в БД", req.TaskId)
		} else {
			log.Printf("gRPC: Ошибка завершения задачи ID %d в БД: %v", req.TaskId, taskErr)
		}
	case *pb.SubmitResultRequest_Error:
		log.Printf("gRPC: Задача ID %d завершилась ошибкой: %s", req.TaskId, result.Error.Message)
		taskErr = s.dbStore.FailTask(req.TaskId)
		if taskErr != nil {
			log.Printf("gRPC: Ошибка отметки задачи ID %d как ошибочной в БД: %v", req.TaskId, taskErr)
		}
	default:
		log.Printf("gRPC: Получен некорректный статус результата для задачи ID %d", req.TaskId)
		return nil, status.Error(codes.InvalidArgument, "некорректный формат статуса результата")
	}

	if taskErr != nil {
		return nil, status.Errorf(codes.Internal, "ошибка БД при обновлении задачи: %v", taskErr)
	}

	go s.scheduler.ProcessTaskCompletion(req.TaskId)

	return &pb.SubmitResultResponse{Acknowledged: true}, nil
}

func (s *grpcServer) getOperationTimeMs(op string) int32 {
	var t int
	switch op {
	case "+":
		t = s.opTimes.Addition
	case "-":
		t = s.opTimes.Subtraction
	case "*":
		t = s.opTimes.Multiplication
	case "/":
		t = s.opTimes.Division
	default:
		t = 1000 // Время по умолчанию
	}
	return int32(t)
}
