package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"room-mate-api/internal/middleware"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/service"
)

func TestTaskHTTPPostgres(t *testing.T) {
	dsn := os.Getenv("TASK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TASK_TEST_DATABASE_URL for API integration tests")
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	must(err)
	pool, err := db.DB()
	must(err)
	defer pool.Close()
	schema := "task_http_" + uuid.New().String()[:8]
	must(db.Exec("CREATE SCHEMA " + schema).Error)
	defer db.Exec("DROP SCHEMA " + schema + " CASCADE")
	scoped, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{})
	must(err)
	scopedPool, err := scoped.DB()
	must(err)
	defer scopedPool.Close()
	files, err := filepath.Glob("../../migrations/*.up.sql")
	must(err)
	for _, file := range files {
		body, err := os.ReadFile(file)
		must(err)
		must(scoped.Exec(string(body)).Error)
	}
	userID, roomID, memberID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	must(scoped.Exec("INSERT INTO users(id,name,email,created_at) VALUES (?, 'Owner', 'owner@test.invalid', ?)", userID, now).Error)
	must(scoped.Exec("INSERT INTO rooms(id,room_name,room_code,room_number,created_by,created_at) VALUES (?, 'Test', 'HTTP', '1', ?, ?)", roomID, userID, now).Error)
	must(scoped.Exec("INSERT INTO room_members(id,room_id,user_id,role,joined_at) VALUES (?,?,?,'owner',?)", memberID, roomID, userID, now).Error)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.AuthMiddleware("test-secret"))
	h := NewTaskHandler(service.NewTaskService(repository.NewTaskRepository(scoped)))
	router.POST("/rooms/:id/tasks", h.Create)
	router.GET("/rooms/:id/tasks/:taskId/occurrences", h.Occurrences)
	router.POST("/rooms/:id/tasks/:taskId/occurrences/:occurrenceId/complete", h.Complete)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: userID.String(), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}).SignedString([]byte("test-secret"))
	must(err)
	request := func(method, path string, body any, authenticated bool) *httptest.ResponseRecorder {
		t.Helper()
		payload, err := json.Marshal(body)
		must(err)
		req := httptest.NewRequest(method, path, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if authenticated {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	base := "/rooms/" + roomID.String() + "/tasks"
	if response := request("POST", base, map[string]any{}, false); response.Code != 401 {
		t.Fatalf("unauthenticated: %d", response.Code)
	}
	response := request("POST", base, map[string]any{"name": "Kitchen", "scheduleType": "one_time", "firstDueAt": now.Add(time.Hour), "timezone": "Asia/Bangkok", "memberIds": []uuid.UUID{memberID}}, true)
	if response.Code != 201 {
		t.Fatalf("create: %d %s", response.Code, response.Body.String())
	}
	var task model.Task
	must(json.Unmarshal(response.Body.Bytes(), &task))
	path := base + "/" + task.ID.String() + "/occurrences"
	response = request("GET", path, nil, true)
	if response.Code != 200 {
		t.Fatalf("list: %s", response.Body.String())
	}
	var result struct {
		Occurrences []model.TaskOccurrence `json:"occurrences"`
	}
	must(json.Unmarshal(response.Body.Bytes(), &result))
	if len(result.Occurrences) != 1 {
		t.Fatal("missing first occurrence")
	}
	response = request("POST", path+"/"+result.Occurrences[0].ID.String()+"/complete", map[string]any{}, true)
	if response.Code != 200 {
		t.Fatalf("complete: %d %s", response.Code, response.Body.String())
	}
	var occurrence model.TaskOccurrence
	must(json.Unmarshal(response.Body.Bytes(), &occurrence))
	if occurrence.Status != "completed" || occurrence.CompletedByMemberID == nil || *occurrence.CompletedByMemberID != memberID {
		t.Fatal("invalid completion response")
	}
}
