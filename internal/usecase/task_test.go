package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ryuta-github/golang-api-server/internal/domain"
)

type fakeTaskRepository struct {
	tasks       map[string]domain.Task
	lastContext context.Context
}

func newFakeTaskRepository() *fakeTaskRepository {
	return &fakeTaskRepository{tasks: make(map[string]domain.Task)}
}

func (f *fakeTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	f.lastContext = ctx
	task.ID = "task-1"
	task.CreatedAt = time.Now()
	task.UpdatedAt = task.CreatedAt
	f.tasks[task.ID] = *task
	return nil
}

func (f *fakeTaskRepository) List(ctx context.Context) ([]domain.Task, error) {
	f.lastContext = ctx
	tasks := make([]domain.Task, 0, len(f.tasks))
	for _, task := range f.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (f *fakeTaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	f.lastContext = ctx
	task, ok := f.tasks[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &task, nil
}

func (f *fakeTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	f.lastContext = ctx
	if _, ok := f.tasks[task.ID]; !ok {
		return domain.ErrNotFound
	}
	f.tasks[task.ID] = *task
	return nil
}

func (f *fakeTaskRepository) Delete(ctx context.Context, id string) error {
	f.lastContext = ctx
	if _, ok := f.tasks[id]; !ok {
		return domain.ErrNotFound
	}
	delete(f.tasks, id)
	return nil
}

func TestTaskCRUD(t *testing.T) {
	t.Parallel()

	repository := newFakeTaskRepository()
	useCase := NewTask(repository)
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("request"), "request-1")

	created, err := useCase.Create(ctx, "first task", "details")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.lastContext != ctx {
		t.Fatal("Create() did not propagate context")
	}

	title := "updated task"
	completed := true
	updated, err := useCase.Update(ctx, created.ID, &title, nil, &completed)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != title || !updated.Completed {
		t.Fatalf("Update() task = %+v", updated)
	}

	if err := useCase.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := useCase.Get(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Get() after delete error = %v, want ErrNotFound", err)
	}
}

func TestTaskUpdateRequiresAField(t *testing.T) {
	t.Parallel()

	useCase := NewTask(newFakeTaskRepository())
	if _, err := useCase.Update(context.Background(), "task-1", nil, nil, nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Update() error = %v, want ErrValidation", err)
	}
}
