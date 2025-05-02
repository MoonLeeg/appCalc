package orchestrator

import (
	"calculator/internal/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HTTPHandlers struct {
	auth      *AuthService
	db        *database.Store
	scheduler *Scheduler // Добавлена зависимость от планировщика
}

func NewHTTPHandlers(auth *AuthService, db *database.Store, scheduler *Scheduler) *HTTPHandlers {
	return &HTTPHandlers{
		auth:      auth,
		db:        db,
		scheduler: scheduler, // Инициализируем планировщик
	}
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *HTTPHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ошибка декодирования запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	login := strings.TrimSpace(req.Login)
	password := strings.TrimSpace(req.Password)

	if login == "" || password == "" {
		http.Error(w, "Логин и пароль не могут быть пустыми", http.StatusBadRequest)
		return
	}

	if len(password) < 6 {
		http.Error(w, "Пароль должен быть не менее 6 символов", http.StatusBadRequest)
		return
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("Ошибка хэширования пароля для пользователя %s: %v", login, err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = h.db.CreateUser(login, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "уже существует") {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
		} else {
			log.Printf("Ошибка создания пользователя %s в БД: %v", login, err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated) // 201 Created
	fmt.Fprintf(w, "Пользователь '%s' успешно зарегистрирован", login)
}

func (h *HTTPHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ошибка декодирования запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	login := strings.TrimSpace(req.Login)
	password := strings.TrimSpace(req.Password)

	if login == "" || password == "" {
		http.Error(w, "Логин и пароль не могут быть пустыми", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUserByLogin(login)
	if err != nil {
		log.Printf("Ошибка получения пользователя %s из БД: %v", login, err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if user == nil || !CheckPasswordHash(password, user.PasswordHash) {
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	tokenString, err := h.auth.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("Ошибка генерации JWT для пользователя %s (ID: %d): %v", login, user.ID, err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{Token: tokenString}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type CalculateRequest struct {
	Expression string `json:"expression"`
}

func (h *HTTPHandlers) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		log.Println("Ошибка: не удалось получить userID из контекста в CalculateHandler")
		http.Error(w, "Внутренняя ошибка сервера (контекст пользователя)", http.StatusInternalServerError)
		return
	}

	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ошибка декодирования JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	exprStr := strings.TrimSpace(req.Expression) // Восстановлено определение exprStr

	if exprStr == "" {
		http.Error(w, "Пустое выражение недопустимо", http.StatusBadRequest)
		return
	}


	exprID, err := h.db.CreateExpression(userID, exprStr) // userID и exprStr теперь определены
	if err != nil {
		log.Printf("Ошибка создания выражения в БД для пользователя %d: %v", userID, err)
		http.Error(w, "Внутренняя ошибка сервера при сохранении выражения", http.StatusInternalServerError)
		return
	}

	log.Printf("Создано выражение ID %d для пользователя %d: %s", exprID, userID, exprStr)

	go func(id int64, expression string) {
		err := h.scheduler.ScheduleTasks(id, expression)
		if err != nil {
			log.Printf("Асинхронная ошибка планирования задач для выражения ID %d: %v", id, err)
		}
	}(exprID, exprStr)

	respData := map[string]interface{}{
		"id":         exprID,
		"expression": exprStr,
		"status":     database.StatusPending, // Начальный статус
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	if err := json.NewEncoder(w).Encode(respData); err != nil {
		log.Printf("Ошибка записи JSON ответа для CalculateHandler (exprID: %d): %v", exprID, err)
	}
}

func EnableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                                                                                   // Разрешаем все источники (для разработки)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")                                                    // Разрешенные методы
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization") // Разрешенные заголовки

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (h *HTTPHandlers) ExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		log.Println("Ошибка: не удалось получить userID из контекста в ExpressionsHandler")
		http.Error(w, "Внутренняя ошибка сервера (контекст пользователя)", http.StatusInternalServerError)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions")
	idStr := strings.Trim(path, "/")

	w.Header().Set("Content-Type", "application/json")

	if idStr == "" { // Запрос списка выражений
		expressions, err := h.db.GetExpressionsByUserID(userID)
		if err != nil {
			log.Printf("Ошибка получения списка выражений для пользователя %d: %v", userID, err)
			http.Error(w, "Внутренняя ошибка сервера при получении выражений", http.StatusInternalServerError)
			return
		}
		if expressions == nil {
			expressions = []database.Expression{}
		}
		if err := json.NewEncoder(w).Encode(expressions); err != nil {
			log.Printf("Ошибка записи JSON ответа для списка выражений (userID: %d): %v", userID, err)
		}
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Неверный ID выражения: "+idStr, http.StatusBadRequest)
		return
	}

	expression, err := h.db.GetExpressionByID(id, userID)
	if err != nil {
		log.Printf("Ошибка получения выражения ID %d для пользователя %d: %v", id, userID, err)
		http.Error(w, "Внутренняя ошибка сервера при получении выражения", http.StatusInternalServerError)
		return
	}

	if expression == nil {
		http.Error(w, fmt.Sprintf("Выражение с ID %d не найдено или доступ запрещен", id), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(expression); err != nil {
		log.Printf("Ошибка записи JSON ответа для выражения ID %d (userID: %d): %v", id, userID, err)
	}
}
