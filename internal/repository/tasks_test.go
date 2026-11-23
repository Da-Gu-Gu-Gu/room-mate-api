package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"room-mate-api/internal/model"
)

func TestNextTaskDue(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	due := time.Date(2026, 3, 7, 20, 0, 0, 0, location)
	next := NextTaskDue(due, "daily", location)
	if next.Hour() != 20 || next.Sub(due) != 23*time.Hour {
		t.Fatalf("DST changed local deadline: %v", next)
	}
	if NextTaskDue(due, "weekly", location).Day() != 14 {
		t.Fatal("weekly recurrence")
	}
}

func TestTaskLifecyclePostgres(t *testing.T) {
	dsn := os.Getenv("TASK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TASK_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	// A unique schema isolates the fixture from all existing data.
	schema := "task_test_" + uuid.New().String()[:8]
	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	defer db.Exec("DROP SCHEMA " + schema + " CASCADE")
	scoped, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	scopedSQL, err := scoped.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer scopedSQL.Close()
	files, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := scoped.Exec(string(body)).Error; err != nil {
			t.Fatalf("%s: %v", file, err)
		}
	}
	ctx := context.Background()
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	roomID := uuid.New()
	users := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	members := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for i, user := range users {
		must(scoped.Exec("INSERT INTO users(id,name,email,created_at) VALUES (?, 'Member', ?, ?)", user, user.String()+"@test.invalid", now).Error)
		role := "member"
		if i == 0 {
			role = "owner"
		}
		must(scoped.Exec("INSERT INTO room_members(id,room_id,user_id,role,joined_at) VALUES (?,?,?,?,?)", members[i], roomID, user, role, now).Error)
	}
	must(scoped.Exec("INSERT INTO rooms(id,room_name,room_code,room_number,created_by,created_at) VALUES (?, 'Test', ?, '1', ?, ?)", roomID, roomID.String(), users[0], now).Error)
	repo := NewTaskRepository(scoped)
	frequency := "daily"
	task := &model.Task{ID: uuid.New(), RoomID: roomID, Name: "Kitchen", ScheduleType: "recurring", CreatedBy: users[0], FirstDueAt: now, Frequency: &frequency, Timezone: "Asia/Bangkok", GracePeriodMinutes: 1440, Status: "active"}
	for i, id := range members {
		task.Members = append(task.Members, model.TaskMember{ID: uuid.New(), TaskID: task.ID, MemberID: id, OrderNo: i + 1})
	}
	must(repo.Create(ctx, task))
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- repo.Sync(ctx, now) }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		must(err)
	}
	var occurrences []model.TaskOccurrence
	must(scoped.Where("task_id = ?", task.ID).Order("due_at").Find(&occurrences).Error)
	if len(occurrences) != 8 {
		t.Fatalf("expected 8 distinct turns, got %d", len(occurrences))
	}
	for i, occurrence := range occurrences {
		if occurrence.AssignedMemberID != members[i%3] {
			t.Fatalf("rotation mismatch at %d", i)
		}
	}
	covered, err := repo.Complete(ctx, roomID, task.ID, occurrences[0].ID, users[1], nil, now.Add(time.Hour))
	must(err)
	if covered.AssignedMemberID != members[0] || *covered.CompletedByMemberID != members[1] {
		t.Fatal("cover lost assignment")
	}
	_, err = repo.Complete(ctx, roomID, task.ID, covered.ID, users[1], nil, now.Add(2*time.Hour))
	must(err)
	_, err = repo.Complete(ctx, roomID, task.ID, covered.ID, users[2], nil, now.Add(2*time.Hour))
	if !errors.Is(err, ErrTaskConflict) {
		t.Fatalf("competing completion: %v", err)
	}
	_, err = repo.Get(ctx, roomID, task.ID, uuid.New())
	if !errors.Is(err, ErrRoomMemberForbidden) {
		t.Fatalf("outsider access: %v", err)
	}
	_, err = repo.Complete(ctx, roomID, task.ID, occurrences[1].ID, users[1], nil, occurrences[1].ExpiresAt)
	if !errors.Is(err, ErrTaskConflict) {
		t.Fatalf("expiry boundary: %v", err)
	}
	var expired model.TaskOccurrence
	must(scoped.First(&expired, "id = ?", occurrences[1].ID).Error)
	if expired.Status != "missed" {
		t.Fatal("expired turn was not persisted")
	}
	must(scoped.Model(&model.RoomMember{}).Where("id = ?", members[2]).Update("leave_at", now).Error)
	must(repo.Sync(ctx, now.Add(48*time.Hour)))
	var newest model.TaskOccurrence
	must(scoped.Where("task_id = ?", task.ID).Order("due_at DESC").First(&newest).Error)
	if newest.AssignedMemberID == members[2] {
		t.Fatal("departed member assigned")
	}
	err = repo.UpdateStatus(ctx, roomID, task.ID, users[1], "paused", now)
	if !errors.Is(err, ErrRoomMemberForbidden) {
		t.Fatal("ordinary member changed schedule")
	}
	must(repo.UpdateStatus(ctx, roomID, task.ID, users[0], "archived", now))
	var count int64
	must(scoped.Model(&model.TaskOccurrence{}).Where("task_id = ? AND status = 'pending'", task.ID).Count(&count).Error)
	if count != 0 {
		t.Fatal("archive left pending turns")
	}
	// One-time tasks generate exactly one turn, even after repeated scheduler runs.
	one := &model.Task{ID: uuid.New(), RoomID: roomID, Name: "Repair", ScheduleType: "one_time", CreatedBy: users[0], FirstDueAt: now.Add(time.Hour), Timezone: "UTC", Status: "active"}
	one.Members = []model.TaskMember{{ID: uuid.New(), TaskID: one.ID, MemberID: members[0], OrderNo: 1}}
	must(repo.Create(ctx, one))
	must(repo.Sync(ctx, now.AddDate(0, 0, 30)))
	must(scoped.Model(&model.TaskOccurrence{}).Where("task_id = ?", one.ID).Count(&count).Error)
	if count != 1 {
		t.Fatal("one-time task repeated")
	}
	// The cutoff includes a deadline equal to repeat_until.
	cutoff := now.AddDate(0, 0, 2)
	finite := &model.Task{ID: uuid.New(), RoomID: roomID, Name: "Finite", ScheduleType: "recurring", CreatedBy: users[0], FirstDueAt: now, Frequency: &frequency, RepeatUntil: &cutoff, Timezone: "UTC", Status: "active"}
	finite.Members = []model.TaskMember{{ID: uuid.New(), TaskID: finite.ID, MemberID: members[0], OrderNo: 1}}
	must(repo.Create(ctx, finite))
	must(repo.Sync(ctx, now.AddDate(0, 0, 30)))
	must(scoped.Model(&model.TaskOccurrence{}).Where("task_id = ?", finite.ID).Count(&count).Error)
	if count != 3 {
		t.Fatalf("inclusive cutoff: %d", count)
	}
	// Verify migration rollback and reapplication in this isolated schema.
	body, err := os.ReadFile("../../migrations/000015_tasks.down.sql")
	must(err)
	must(scoped.Exec(string(body)).Error)
	body, err = os.ReadFile("../../migrations/000015_tasks.up.sql")
	must(err)
	must(scoped.Exec(string(body)).Error)
}
