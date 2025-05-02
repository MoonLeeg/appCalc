package database

import (
	"strings"
	"testing"
)

func TestUserAndExpressionCRUD(t *testing.T) {
	store, err := NewStore(":memory:")
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skipf("skip DB tests: %v", err)
		}
		t.Fatalf("NewStore error: %v", err)
	}
	if err := store.InitDB(); err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skipf("skip DB tests: %v", err)
		}
		t.Fatalf("InitDB error: %v", err)
	}

	uid, err := store.CreateUser("testuser", "hashpass")
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	user, err := store.GetUserByLogin("testuser")
	if err != nil {
		t.Fatalf("GetUserByLogin error: %v", err)
	}
	if user == nil || user.ID != uid {
		t.Fatalf("GetUserByLogin returned wrong user: %+v", user)
	}

	exprID, err := store.CreateExpression(uid, "1+1")
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}
	expr, err := store.GetExpressionByID(exprID, uid)
	if err != nil {
		t.Fatalf("GetExpressionByID error: %v", err)
	}
	if expr == nil || expr.Expression != "1+1" || expr.Status != StatusPending {
		t.Fatalf("GetExpressionByID returned wrong expression: %+v", expr)
	}

	list, err := store.GetExpressionsByUserID(uid)
	if err != nil {
		t.Fatalf("GetExpressionsByUserID error: %v", err)
	}
	if len(list) != 1 || list[0].ID != exprID {
		t.Fatalf("GetExpressionsByUserID returned wrong list: %+v", list)
	}
}

func TestTaskLifecycle(t *testing.T) {
	store, err := NewStore(":memory:")
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skipf("skip DB tests: %v", err)
		}
		t.Fatalf("NewStore error: %v", err)
	}
	if err := store.InitDB(); err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skipf("skip DB tests: %v", err)
		}
		t.Fatalf("InitDB error: %v", err)
	}

	uid, err := store.CreateUser("u2", "h2")
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	exprID, err := store.CreateExpression(uid, "2*3")
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	tid, err := store.CreateTask(exprID, "*", 2, 3)
	if err != nil {
		t.Fatalf("CreateTask error: %v", err)
	}

	task, err := store.GetAndLeasePendingTask()
	if err != nil {
		t.Fatalf("GetAndLeasePendingTask error: %v", err)
	}
	if task == nil || task.ID != tid || task.Status != StatusInProgress {
		t.Fatalf("GetAndLeasePendingTask returned wrong: %+v", task)
	}

	if err := store.CompleteTask(tid, 6); err != nil {
		t.Fatalf("CompleteTask error: %v", err)
	}
	t2, err := store.GetTaskByID(tid)
	if err != nil {
		t.Fatalf("GetTaskByID error: %v", err)
	}
	if t2.Status != StatusDone || !t2.Result.Valid || t2.Result.Float64 != 6 {
		t.Fatalf("GetTaskByID after complete wrong: %+v", t2)
	}

	tid2, err := store.CreateTask(exprID, "+", 1, 1)
	if err != nil {
		t.Fatalf("CreateTask error: %v", err)
	}
	task2, _ := store.GetAndLeasePendingTask()
	if task2 == nil {
		t.Fatalf("Expected task2 leased, got nil")
	}
	if err := store.FailTask(tid2); err != nil {
		t.Fatalf("FailTask error: %v", err)
	}
	t3, _ := store.GetTaskByID(tid2)
	if t3.Status != StatusPending || t3.Retries != 1 {
		t.Fatalf("FailTask not applied: %+v", t3)
	}

	has, err := store.HasPendingTasks(exprID)
	if err != nil {
		t.Fatalf("HasPendingTasks error: %v", err)
	}
	if !has {
		t.Fatalf("HasPendingTasks returned false, expected true")
	}

	has2, _ := store.HasPendingTasks(9999)
	if has2 {
		t.Fatalf("HasPendingTasks for unknown expr should be false")
	}
}