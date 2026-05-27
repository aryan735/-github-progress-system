package handler

import (
	"net/http"

	"github.com/aryan735/-github-progress-system/internal/github"
	"github.com/aryan735/-github-progress-system/internal/scheduler"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	client  *github.Client
	daily   *scheduler.DailyService
}

func New(client *github.Client, daily *scheduler.DailyService) *Handler {
	return &Handler{
		client: client,
		daily:  daily,
	}
}

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/health", h.Health)
	e.GET("/activity/today", h.ActivityToday)
	e.POST("/progress/daily", h.ProgressDaily)
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) ActivityToday(c echo.Context) error {
	summary, err := h.client.CollectTodayCommits(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, summary)
}

func (h *Handler) ProgressDaily(c echo.Context) error {
	result, err := h.daily.Run(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
