package domain

import "testing"

func TestNewTask(t *testing.T) {
	t.Parallel()

	task, err := NewTask("  write tests  ", "description")
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}
	if task.Title != "write tests" {
		t.Fatalf("Title = %q, want %q", task.Title, "write tests")
	}
}

func TestNewTaskRejectsInvalidTitle(t *testing.T) {
	t.Parallel()

	if _, err := NewTask("   ", ""); err != ErrValidation {
		t.Fatalf("NewTask() error = %v, want %v", err, ErrValidation)
	}
}

func TestTaskUpdate(t *testing.T) {
	t.Parallel()

	task, _ := NewTask("before", "")
	title := "after"
	completed := true
	if err := task.Update(&title, nil, &completed); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if task.Title != title || !task.Completed {
		t.Fatalf("Update() task = %+v", task)
	}
}
