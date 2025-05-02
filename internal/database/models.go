package database

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"` // Не отправляем хэш клиенту
	CreatedAt    time.Time `json:"created_at"`
}

type Expression struct {
	ID         int64           `json:"id"`
	UserID     int64           `json:"user_id"`
	Expression string          `json:"expression"`
	Status     string          `json:"status"`           // pending, in_progress, done, error
	Result     sql.NullFloat64 `json:"result,omitempty"` // Используем NullFloat64 для поддержки NULL в БД
	Steps      sql.NullString  `json:"steps,omitempty"`  // Шаги можно хранить как JSON строку
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Task struct {
	ID           int64           `json:"id"`
	ExpressionID int64           `json:"expression_id"`
	Operation    string          `json:"operation"` // +, -, *, /
	Arg1         float64         `json:"arg1"`
	Arg2         float64         `json:"arg2"`
	Result       sql.NullFloat64 `json:"result,omitempty"`
	Status       string          `json:"status"` // pending, in_progress, done, error
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Retries      int             `json:"retries"`
}

const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
	StatusError      = "error"
)
