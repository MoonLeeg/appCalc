package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type Store struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL") // Включаем foreign keys и WAL
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия БД %s: %w", dbPath, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка подключения к БД %s: %w", dbPath, err)
	}

	store := &Store{
		db:   db,
		path: dbPath,
	}

	log.Printf("Успешное подключение к базе данных: %s", dbPath)
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) InitDB() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS expressions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			expression TEXT NOT NULL,
			status TEXT NOT NULL,
			result REAL,
			steps TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			expression_id INTEGER NOT NULL,
			operation TEXT NOT NULL,
			arg1 REAL NOT NULL,
			arg2 REAL NOT NULL,
			result REAL,
			status TEXT NOT NULL,
			retries INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(expression_id) REFERENCES expressions(id)
		)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("database migration error: %w", err)
		}
	}
	return nil
}

func (s *Store) CreateUser(login, passwordHash string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO users (login, password_hash) VALUES (?, ?)`
	res, err := s.db.Exec(query, login, passwordHash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.login") {
			return 0, fmt.Errorf("пользователь с логином '%s' уже существует", login)
		}
		return 0, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID нового пользователя: %w", err)
	}

	log.Printf("Создан пользователь '%s' с ID: %d", login, id)
	return id, nil
}

func (s *Store) GetUserByLogin(login string) (*User, error) {
	s.mu.RLock() // Используем RLock для чтения
	defer s.mu.RUnlock()

	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = ?`
	row := s.db.QueryRow(query, login)

	user := &User{}
	err := row.Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Пользователь не найден - это не ошибка для этой функции
		}
		return nil, fmt.Errorf("ошибка поиска пользователя по логину '%s': %w", login, err)
	}

	return user, nil
}

func (s *Store) CreateExpression(userID int64, expression string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO expressions (user_id, expression, status) VALUES (?, ?, ?)`
	res, err := s.db.Exec(query, userID, expression, StatusPending)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания выражения: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID нового выражения: %w", err)
	}

	log.Printf("Создано выражение ID %d для пользователя ID %d: %s", id, userID, expression)
	return id, nil
}

func (s *Store) GetExpressionByID(id, userID int64) (*Expression, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, user_id, expression, status, result, steps, created_at, updated_at
	         FROM expressions WHERE id = ? AND user_id = ?`
	row := s.db.QueryRow(query, id, userID)

	expr := &Expression{}
	err := row.Scan(
		&expr.ID, &expr.UserID, &expr.Expression, &expr.Status,
		&expr.Result, &expr.Steps, &expr.CreatedAt, &expr.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Не найдено
		}
		return nil, fmt.Errorf("ошибка получения выражения ID %d: %w", id, err)
	}
	return expr, nil
}

func (s *Store) GetExpressionsByUserID(userID int64) ([]Expression, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, user_id, expression, status, result, steps, created_at, updated_at
	         FROM expressions WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка выражений для пользователя ID %d: %w", userID, err)
	}
	defer rows.Close()

	var expressions []Expression
	for rows.Next() {
		expr := Expression{}
		err := rows.Scan(
			&expr.ID, &expr.UserID, &expr.Expression, &expr.Status,
			&expr.Result, &expr.Steps, &expr.CreatedAt, &expr.UpdatedAt,
		)
		if err != nil {
			log.Printf("Ошибка сканирования строки выражения: %v", err)
			continue
		}
		expressions = append(expressions, expr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по списку выражений: %w", err)
	}

	return expressions, nil
}

func (s *Store) UpdateExpressionStatusResult(id int64, status string, result sql.NullFloat64, stepsJSON sql.NullString) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `UPDATE expressions SET status = ?, result = ?, steps = ?, updated_at = CURRENT_TIMESTAMP
	         WHERE id = ?`
	_, err := s.db.Exec(query, status, result, stepsJSON, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления выражения ID %d: %w", id, err)
	}
	log.Printf("Обновлен статус/результат выражения ID %d: Статус=%s", id, status)
	return nil
}

func (s *Store) CreateTask(expressionID int64, operation string, arg1, arg2 float64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO tasks (expression_id, operation, arg1, arg2, status) VALUES (?, ?, ?, ?, ?)`
	res, err := s.db.Exec(query, expressionID, operation, arg1, arg2, StatusPending)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания задачи для выражения ID %d: %w", expressionID, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID новой задачи: %w", err)
	}

	log.Printf("Создана задача ID %d для выражения ID %d: %f %s %f", id, expressionID, arg1, operation, arg2)
	return id, nil
}

func (s *Store) GetAndLeasePendingTask() (*Task, error) {
	s.mu.Lock() // Используем полную блокировку, так как чтение и запись
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции для получения задачи: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Восстанавливаем панику
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Printf("Ошибка коммита транзакции при получении задачи: %v", err)
			}
		}
	}()

	querySelect := `SELECT id, expression_id, operation, arg1, arg2, status, retries, created_at, updated_at
	                FROM tasks WHERE status = ? ORDER BY created_at ASC LIMIT 1`
	row := tx.QueryRow(querySelect, StatusPending)

	task := &Task{}
	err = row.Scan(
		&task.ID, &task.ExpressionID, &task.Operation, &task.Arg1, &task.Arg2,
		&task.Status, &task.Retries, &task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет ожидающих задач
		}
		return nil, fmt.Errorf("ошибка поиска ожидающей задачи: %w", err)
	}

	queryUpdate := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = tx.Exec(queryUpdate, StatusInProgress, task.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса задачи ID %d: %w", task.ID, err)
	}

	task.Status = StatusInProgress // Обновляем статус в возвращаемом объекте
	log.Printf("Задача ID %d взята в обработку (Expression ID: %d)", task.ID, task.ExpressionID)
	return task, nil // err будет nil здесь, defer обработает Commit
}

func (s *Store) CompleteTask(taskID int64, result float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `UPDATE tasks SET status = ?, result = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = ?`
	res, err := s.db.Exec(query, StatusDone, result, taskID, StatusInProgress)
	if err != nil {
		return fmt.Errorf("ошибка завершения задачи ID %d: %w", taskID, err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Предупреждение: Попытка завершить задачу ID %d, которая не найдена или уже не в статусе '%s'", taskID, StatusInProgress)
	}

	log.Printf("Задача ID %d завершена с результатом: %f", taskID, result)
	return nil
}

func (s *Store) FailTask(taskID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `UPDATE tasks SET status = ?, retries = retries + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = ?`
	res, err := s.db.Exec(query, StatusPending, taskID, StatusInProgress)
	if err != nil {
		return fmt.Errorf("ошибка отметки задачи ID %d как ошибочной: %w", taskID, err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Предупреждение: Попытка отметить ошибку для задачи ID %d, которая не найдена или уже не в статусе '%s'", taskID, StatusInProgress)
	}

	log.Printf("Ошибка выполнения задачи ID %d, возвращена в очередь.", taskID)
	return nil
}

func (s *Store) GetTaskByID(taskID int64) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, expression_id, operation, arg1, arg2, result, status, retries, created_at, updated_at
	         FROM tasks WHERE id = ?`
	row := s.db.QueryRow(query, taskID)

	task := &Task{}
	err := row.Scan(
		&task.ID, &task.ExpressionID, &task.Operation, &task.Arg1, &task.Arg2,
		&task.Result, &task.Status, &task.Retries, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Не найдено
		}
		return nil, fmt.Errorf("ошибка получения задачи ID %d: %w", taskID, err)
	}
	return task, nil
}

func (s *Store) HasPendingTasks(expressionID int64) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT 1 FROM tasks WHERE expression_id = ? AND status IN (?, ?) LIMIT 1`
	var exists int
	err := s.db.QueryRow(query, expressionID, StatusPending, StatusInProgress).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Нет незавершенных задач
		}
		return false, fmt.Errorf("ошибка проверки незавершенных задач для выражения ID %d: %w", expressionID, err)
	}
	return true, nil // Найдена хотя бы одна незавершенная задача
}

func (s *Store) GetExpressionByIDInternal(id int64) (*Expression, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, user_id, expression, status, result, steps, created_at, updated_at
	         FROM expressions WHERE id = ?`
	row := s.db.QueryRow(query, id)

	expr := &Expression{}
	err := row.Scan(
		&expr.ID, &expr.UserID, &expr.Expression, &expr.Status,
		&expr.Result, &expr.Steps, &expr.CreatedAt, &expr.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Не найдено
		}
		return nil, fmt.Errorf("ошибка получения выражения ID %d (внутр.): %w", id, err)
	}
	return expr, nil
}

// GetAllTasksForExpression возвращает все задачи для данного выражения.
func (s *Store) GetAllTasksForExpression(expressionID int64) ([]Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, expression_id, operation, arg1, arg2, result, status, retries, created_at, updated_at
		FROM tasks WHERE expression_id = ?`
	rows, err := s.db.Query(query, expressionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса задач для выражения ID %d: %w", expressionID, err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(
			&task.ID, &task.ExpressionID, &task.Operation,
			&task.Arg1, &task.Arg2, &task.Result,
			&task.Status, &task.Retries, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			log.Printf("Ошибка сканирования строки задачи при GetAllTasksForExpression: %v", err)
			continue // Пропускаем ошибочную строку, но продолжаем с остальными
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по строкам задач: %w", err)
	}

	return tasks, nil
}
