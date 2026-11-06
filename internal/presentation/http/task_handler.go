package http

import (
	"errors"
	stdhttp "net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ryuta-github/golang-api-server/internal/domain"
	"github.com/ryuta-github/golang-api-server/internal/usecase"
)

type TaskHandler struct {
	usecase *usecase.Task
}

func NewTaskHandler(taskUseCase *usecase.Task) *TaskHandler {
	return &TaskHandler{usecase: taskUseCase}
}

func (h *TaskHandler) Register(group *echo.Group) {
	group.POST("", h.create)
	group.GET("", h.list)
	group.GET("/:id", h.get)
	group.PATCH("/:id", h.update)
	group.DELETE("/:id", h.delete)
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

type taskResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *TaskHandler) create(c echo.Context) error {
	var request createTaskRequest
	if err := c.Bind(&request); err != nil {
		return writeError(c, domain.ErrValidation)
	}
	task, err := h.usecase.Create(c.Request().Context(), request.Title, request.Description)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(stdhttp.StatusCreated, toTaskResponse(task))
}

func (h *TaskHandler) list(c echo.Context) error {
	tasks, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	response := make([]taskResponse, len(tasks))
	for i := range tasks {
		response[i] = toTaskResponse(&tasks[i])
	}
	return c.JSON(stdhttp.StatusOK, response)
}

func (h *TaskHandler) get(c echo.Context) error {
	task, err := h.usecase.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(stdhttp.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) update(c echo.Context) error {
	var request updateTaskRequest
	if err := c.Bind(&request); err != nil {
		return writeError(c, domain.ErrValidation)
	}
	task, err := h.usecase.Update(c.Request().Context(), c.Param("id"), request.Title, request.Description, request.Completed)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(stdhttp.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) delete(c echo.Context) error {
	if err := h.usecase.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return writeError(c, err)
	}
	return c.NoContent(stdhttp.StatusNoContent)
}

func toTaskResponse(task *domain.Task) taskResponse {
	return taskResponse{
		ID: task.ID, Title: task.Title, Description: task.Description,
		Completed: task.Completed, CreatedAt: task.CreatedAt, UpdatedAt: task.UpdatedAt,
	}
}

func writeError(c echo.Context, err error) error {
	status := stdhttp.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, domain.ErrValidation):
		status, message = stdhttp.StatusBadRequest, "invalid request"
	case errors.Is(err, domain.ErrNotFound):
		status, message = stdhttp.StatusNotFound, "task not found"
	case errors.Is(err, domain.ErrConflict):
		status, message = stdhttp.StatusConflict, "task conflict"
	}
	return c.JSON(status, map[string]string{"error": message})
}
