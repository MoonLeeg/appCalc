package orchestrator

import (
	"calculator/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

type OperationTimes struct {
	Addition       int
	Subtraction    int
	Multiplication int
	Division       int
}

type Scheduler struct {
	dbStore *database.Store
	opTimes *OperationTimes
}

func NewScheduler(db *database.Store) *Scheduler {
	return &Scheduler{
		dbStore: db,
		opTimes: initOperationTimes(), // Инициализируем время операций
	}
}

func (s *Scheduler) ScheduleTasks(expressionID int64, expression string) error {
	parser := NewParser(expression)
	ast, err := parser.Parse()
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка парсинга: %v", err)
		s.dbStore.UpdateExpressionStatusResult(expressionID, database.StatusError, sql.NullFloat64{}, sql.NullString{String: errMsg, Valid: true})
		return fmt.Errorf("ошибка парсинга выражения ID %d: %w", expressionID, err)
	}

	log.Printf("AST для выражения ID %d построено. Начинаем планирование задач.", expressionID)

	err = s.planTasksRecursive(ast, expressionID)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка планирования задач: %v", err)
		s.dbStore.UpdateExpressionStatusResult(expressionID, database.StatusError, sql.NullFloat64{}, sql.NullString{String: errMsg, Valid: true})
		return fmt.Errorf("ошибка планирования задач для выражения ID %d: %w", expressionID, err)
	}

	if ast.Value == nil { // Только если это не просто число
		err = s.dbStore.UpdateExpressionStatusResult(expressionID, database.StatusInProgress, sql.NullFloat64{}, sql.NullString{})
		if err != nil {
			log.Printf("Ошибка обновления статуса на in_progress для выражения ID %d: %v", expressionID, err)
		}
	} else {
		log.Printf("Выражение ID %d является числом (%f), завершаем сразу.", expressionID, *ast.Value)
		stepsJSON, _ := json.Marshal([]string{fmt.Sprintf("Result: %f", *ast.Value)})
		err = s.dbStore.UpdateExpressionStatusResult(expressionID,
			database.StatusDone,
			sql.NullFloat64{Float64: *ast.Value, Valid: true},
			sql.NullString{String: string(stepsJSON), Valid: true},
		)
		if err != nil {
			log.Printf("Ошибка обновления статуса на done для числового выражения ID %d: %v", expressionID, err)
		}
	}

	log.Printf("Планирование задач для выражения ID %d завершено.", expressionID)
	return nil
}

func (s *Scheduler) planTasksRecursive(node *Node, expressionID int64) error {
	if node == nil || node.Value != nil { // Базовый случай: лист (число) или пустой узел
		return nil
	}

	if err := s.planTasksRecursive(node.Left, expressionID); err != nil {
		return err
	}
	if err := s.planTasksRecursive(node.Right, expressionID); err != nil {
		return err
	}

	leftReady := node.Left != nil && node.Left.Value != nil
	rightReady := node.Right != nil && node.Right.Value != nil

	if leftReady && rightReady {
		_, err := s.dbStore.CreateTask(
			expressionID,
			node.Op,
			*node.Left.Value,
			*node.Right.Value,
		)
		if err != nil {
			return fmt.Errorf("ошибка создания задачи для операции '%s' выражения ID %d: %w", node.Op, expressionID, err)
		}
	}

	return nil
}

func (s *Scheduler) GetOperationTimes() *OperationTimes {
	return s.opTimes
}

const maxRetries = 3

// Добавляю helper для заполнения AST результатами задач
func fillASTValues(node *Node, doneTasks []database.Task) {
	if node == nil {
		return
	}
	if node.Value != nil {
		// Значение уже известно (либо лист, либо ранее заполненный узел)
		return
	}
	// Заполняем дочерние узлы сначала
	fillASTValues(node.Left, doneTasks)
	fillASTValues(node.Right, doneTasks)
	// Если оба дочерних значения известны, ищем соответствующую задачу
	if node.Op != "" && node.Left != nil && node.Left.Value != nil && node.Right != nil && node.Right.Value != nil {
		for _, t := range doneTasks {
			if t.Operation == node.Op && t.Arg1 == *node.Left.Value && t.Arg2 == *node.Right.Value {
				val := t.Result.Float64
				node.Value = &val
				break
			}
		}
	}
}

// ProcessTaskCompletion обрабатывает результаты задач, обновляет AST и поочередно создает новые задачи
func (s *Scheduler) ProcessTaskCompletion(taskID int64) {
	log.Printf("Scheduler: Обработка завершения/ошибки задачи ID %d", taskID)

	// Получаем задачу
	task, err := s.dbStore.GetTaskByID(taskID)
	if err != nil {
		log.Printf("Scheduler: Ошибка получения задачи ID %d из БД: %v", taskID, err)
		return
	}
	if task == nil {
		log.Printf("Scheduler: Задача ID %d не найдена", taskID)
		return
	}

	// Получаем выражение
	expr, err := s.dbStore.GetExpressionByIDInternal(task.ExpressionID)
	if err != nil {
		log.Printf("Scheduler: Ошибка получения выражения ID %d из БД: %v", task.ExpressionID, err)
		return
	}
	if expr == nil {
		log.Printf("Scheduler: Выражение ID %d для задачи ID %d не найдено", task.ExpressionID, taskID)
		return
	}

	// Парсим AST выражения
	parser := NewParser(expr.Expression)
	ast, err := parser.Parse()
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка парсинга выражения при обработке задачи ID %d: %v", taskID, err)
		log.Printf("Scheduler: %s", errMsg)
		s.dbStore.UpdateExpressionStatusResult(expr.ID,
			database.StatusError,
			sql.NullFloat64{},
			sql.NullString{String: errMsg, Valid: true},
		)
		return
	}

	// Получаем все завершенные задачи для этого выражения
	allTasks, err := s.dbStore.GetAllTasksForExpression(expr.ID)
	if err != nil {
		log.Printf("Scheduler: Ошибка получения задач для выражения ID %d: %v", expr.ID, err)
	}
	var doneTasks []database.Task
	for _, t := range allTasks {
		if t.Status == database.StatusDone {
			doneTasks = append(doneTasks, t)
		}
	}

	// Заполняем AST значениями из выполненных задач
	fillASTValues(ast, doneTasks)

	// Планируем следующие задачи
	err = s.planTasksRecursive(ast, expr.ID)
	if err != nil {
		log.Printf("Scheduler: Ошибка планирования задач для выражения ID %d: %v", expr.ID, err)
		return
	}

	// Если значение корня AST установлено, выражение завершено
	if ast.Value != nil {
		result := *ast.Value
		s.dbStore.UpdateExpressionStatusResult(expr.ID,
			database.StatusDone,
			sql.NullFloat64{Float64: result, Valid: true},
			sql.NullString{},
		)
		log.Printf("Scheduler: Выражение ID %d успешно завершено с результатом %f.", expr.ID, result)
	} else {
		// Иначе выражение всё ещё в процессе
		s.dbStore.UpdateExpressionStatusResult(expr.ID,
			database.StatusInProgress,
			sql.NullFloat64{},
			sql.NullString{},
		)
	}
}

func initOperationTimes() *OperationTimes {
	return &OperationTimes{
		Addition:       readTimeEnv("TIME_ADDITION_MS", 1000),
		Subtraction:    readTimeEnv("TIME_SUBTRACTION_MS", 1000),
		Multiplication: readTimeEnv("TIME_MULTIPLICATION_MS", 1000),
		Division:       readTimeEnv("TIME_DIVISION_MS", 1000),
	}
}

func readTimeEnv(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if t, err := strconv.Atoi(v); err == nil && t >= 0 {
			return t
		} else {
			fmt.Printf("Предупреждение: Неверное значение для %s ('%s'), используется значение по умолчанию %d\n", key, v, defaultValue)
		}
	}
	return defaultValue
}
