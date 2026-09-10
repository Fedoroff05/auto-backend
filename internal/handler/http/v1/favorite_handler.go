package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Fedoroff05/auto-backend/internal/handler/http/middleware"
	"github.com/Fedoroff05/auto-backend/internal/handler/http/response"
	"github.com/Fedoroff05/auto-backend/internal/usecase"
)

type FavoriteHandler struct {
	favUsecase *usecase.FavoriteUsecase
}

func NewFavoriteHandler(favUsecase *usecase.FavoriteUsecase) *FavoriteHandler {
	return &FavoriteHandler{favUsecase: favUsecase}
}

// Add godoc
// @Summary      Добавить авто в избранное
// @Tags         favorites
// @Security     BearerAuth
// @Param        listing_id path string true "ID объявления"
// @Success      200 {object} response.Response{data=string}
// @Router       /favorites/{listing_id} [post]
func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	listingID, err := uuid.Parse(chi.URLParam(r, "listing_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid listing id")
		return
	}

	if err := h.favUsecase.AddToFavorites(r.Context(), userID, listingID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to add to favorites")
		return
	}

	response.JSON(w, http.StatusOK, "added to favorites")
}

// Remove godoc
// @Summary      Удалить авто из избранного
// @Tags         favorites
// @Security     BearerAuth
// @Param        listing_id path string true "ID объявления"
// @Success      200 {object} response.Response{data=string}
// @Router       /favorites/{listing_id} [delete]
func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	listingID, err := uuid.Parse(chi.URLParam(r, "listing_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid listing id")
		return
	}

	if err := h.favUsecase.RemoveFromFavorites(r.Context(), userID, listingID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to remove from favorites")
		return
	}

	response.JSON(w, http.StatusOK, "removed from favorites")
}

// GetUserFavorites godoc
// @Summary      Получить список избранных авто пользователя
// @Tags         favorites
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=[]domain.Listing}
// @Router       /favorites [get]
func (h *FavoriteHandler) GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	list, err := h.favUsecase.GetUserFavorites(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get favorites")
		return
	}

	response.JSON(w, http.StatusOK, list)
}
