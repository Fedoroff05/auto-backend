package v1

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/Fedoroff05/auto-backend/internal/handler/http/middleware"
	"github.com/Fedoroff05/auto-backend/internal/handler/http/response"
	"github.com/Fedoroff05/auto-backend/internal/usecase"
)

type ABHandler struct {
	abUsecase *usecase.ABUsecase
}

func NewABHandler(abUsecase *usecase.ABUsecase) *ABHandler {
	return &ABHandler{abUsecase: abUsecase}
}

type trackEventRequest struct {
	ExperimentName string          `json:"experiment_name"`
	Variant        string          `json:"variant"`
	EventType      string          `json:"event_type"`
	Payload        json.RawMessage `json:"payload,omitempty"`
}

// GetVariant godoc
// @Summary      Получить вариант A/B теста для текущего пользователя
// @Tags         ab-testing
// @Security     BearerAuth
// @Param        experiment query string true "Название эксперимента"
// @Success      200 {object} response.Response{data=string}
// @Router       /ab/variant [get]
func (h *ABHandler) GetVariant(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	expName := r.URL.Query().Get("experiment")
	if expName == "" {
		response.Error(w, http.StatusBadRequest, "experiment param is required")
		return
	}

	variant := h.abUsecase.GetVariant(userID, expName)
	response.JSON(w, http.StatusOK, map[string]string{
		"experiment": expName,
		"variant":    variant,
	})
}

// TrackEvent godoc
// @Summary      Залогировать аналитическое событие пользователя
// @Tags         ab-testing
// @Accept       json
// @Produce      json
// @Param        request body trackEventRequest true "Событие аналитики"
// @Success      200 {object} response.Response{data=string}
// @Router       /ab/events [post]
func (h *ABHandler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	var req trackEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var userIDPtr *uuid.UUID
	if userID, ok := middleware.GetUserIDFromContext(r.Context()); ok {
		userIDPtr = &userID
	}

	err := h.abUsecase.LogEvent(r.Context(), userIDPtr, req.ExperimentName, req.Variant, req.EventType, req.Payload)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to track event")
		return
	}

	response.JSON(w, http.StatusOK, "event recorded")
}
