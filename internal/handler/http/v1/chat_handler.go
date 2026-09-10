package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Fedoroff05/auto-backend/internal/handler/http/middleware"
	"github.com/Fedoroff05/auto-backend/internal/handler/http/response"
	"github.com/Fedoroff05/auto-backend/internal/handler/ws"
	"github.com/Fedoroff05/auto-backend/internal/usecase"
	"github.com/Fedoroff05/auto-backend/pkg/jwt"
)

type ChatHandler struct {
	chatUsecase  *usecase.ChatUsecase
	hub          *ws.Hub
	tokenManager *jwt.TokenManager
}

func NewChatHandler(chatUsecase *usecase.ChatUsecase, hub *ws.Hub, tokenManager *jwt.TokenManager) *ChatHandler {
	return &ChatHandler{
		chatUsecase:  chatUsecase,
		hub:          hub,
		tokenManager: tokenManager,
	}
}

type startChatRequest struct {
	ListingID uuid.UUID `json:"listing_id"`
}

// StartChat godoc
// @Summary      Создать или получить существующий чат по объявлению
// @Tags         chats
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body startChatRequest true "ID объявления"
// @Success      200 {object} response.Response{data=domain.Chat}
// @Router       /chats/start [post]
func (h *ChatHandler) StartChat(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req startChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	chat, err := h.chatUsecase.StartChat(r.Context(), userID, req.ListingID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, chat)
}

// GetUserChats godoc
// @Summary      Список всех чатов пользователя
// @Tags         chats
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} response.Response{data=[]domain.Chat}
// @Router       /chats [get]
func (h *ChatHandler) GetUserChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chats, err := h.chatUsecase.GetUserChats(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch chats")
		return
	}

	response.JSON(w, http.StatusOK, chats)
}

// GetMessages godoc
// @Summary      История сообщений чата
// @Tags         chats
// @Security     BearerAuth
// @Produce      json
// @Param        chat_id path string true "ID чата"
// @Param        limit query int false "Лимит"
// @Param        offset query int false "Смещение"
// @Success      200 {object} response.Response{data=[]domain.Message}
// @Router       /chats/{chat_id}/messages [get]
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chatID, err := uuid.Parse(chi.URLParam(r, "chat_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	messages, err := h.chatUsecase.GetMessages(r.Context(), userID, chatID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, messages)
}

// устанавливает WebSocket-соединение (токен передается query-параметром ?token=...)
func (h *ChatHandler) ConnectWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token param", http.StatusUnauthorized)
		return
	}

	claims, err := h.tokenManager.ParseAccessToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	ws.ServeWS(h.hub, h.chatUsecase, claims.UserID, w, r)
}
