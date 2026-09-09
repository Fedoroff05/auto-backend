package v1

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Fedoroff05/auto-backend/internal/domain"
	"github.com/Fedoroff05/auto-backend/internal/handler/http/middleware"
	"github.com/Fedoroff05/auto-backend/internal/handler/http/response"
)

type ListingHandler struct {
	listingUsecase domain.ListingUsecase
}

func NewListingHandler(listingUsecase domain.ListingUsecase) *ListingHandler {
	return &ListingHandler{listingUsecase: listingUsecase}
}

// DTO для создания и обновления
type createListingRequest struct {
	BrandID     int     `json:"brand_id"`
	ModelID     int     `json:"model_id"`
	Year        int     `json:"year"`
	Price       float64 `json:"price"`
	Mileage     int     `json:"mileage"`
	VIN         *string `json:"vin,omitempty"`
	Description string  `json:"description"`
}

type listResponse struct {
	Total    int              `json:"total"`
	Listings []domain.Listing `json:"listings"`
}

// Create godoc
// @Summary      Создание объявления
// @Tags         listings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body createListingRequest true "Данные авто"
// @Success      201 {object} response.Response{data=domain.Listing}
// @Router       /listings [post]
func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	listing := &domain.Listing{
		BrandID:     req.BrandID,
		ModelID:     req.ModelID,
		Year:        req.Year,
		Price:       req.Price,
		Mileage:     req.Mileage,
		VIN:         req.VIN,
		Description: req.Description,
	}

	created, err := h.listingUsecase.CreateListing(r.Context(), userID, listing)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create listing")
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

// GetByID godoc
// @Summary      Получение объявления по ID
// @Tags         listings
// @Produce      json
// @Param        id path string true "ID объявления"
// @Success      200 {object} response.Response{data=domain.Listing}
// @Router       /listings/{id} [get]
func (h *ListingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid listing id")
		return
	}

	listing, err := h.listingUsecase.GetListingByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrListingNotFound) {
			response.Error(w, http.StatusNotFound, "listing not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get listing")
		return
	}

	response.JSON(w, http.StatusOK, listing)
}

// List godoc
// @Summary      Поиск по каталогу объявлений с фильтрами
// @Tags         listings
// @Produce      json
// @Param        brand_id query int false "ID марки"
// @Param        model_id query int false "ID модели"
// @Param        min_price query number false "Цена от"
// @Param        max_price query number false "Цена до"
// @Param        min_year query int false "Год от"
// @Param        max_year query int false "Год до"
// @Param        limit query int false "Лимит на страницу"
// @Param        offset query int false "Смещение"
// @Success      200 {object} response.Response{data=listResponse}
// @Router       /listings [get]
func (h *ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.ListingFilter{}

	if val, err := strconv.Atoi(q.Get("brand_id")); err == nil {
		filter.BrandID = &val
	}
	if val, err := strconv.Atoi(q.Get("model_id")); err == nil {
		filter.ModelID = &val
	}
	if val, err := strconv.ParseFloat(q.Get("min_price"), 64); err == nil {
		filter.MinPrice = &val
	}
	if val, err := strconv.ParseFloat(q.Get("max_price"), 64); err == nil {
		filter.MaxPrice = &val
	}
	if val, err := strconv.Atoi(q.Get("min_year")); err == nil {
		filter.MinYear = &val
	}
	if val, err := strconv.Atoi(q.Get("max_year")); err == nil {
		filter.MaxYear = &val
	}
	if val, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = val
	}
	if val, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = val
	}

	listings, total, err := h.listingUsecase.GetListings(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch listings")
		return
	}

	response.JSON(w, http.StatusOK, listResponse{
		Total:    total,
		Listings: listings,
	})
}

// Update godoc
// @Summary      Обновление объявления
// @Tags         listings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "ID объявления"
// @Param        request body createListingRequest true "Новые данные"
// @Success      200 {object} response.Response{data=domain.Listing}
// @Router       /listings/{id} [put]
func (h *ListingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roleVal := r.Context().Value(middleware.UserRoleKey)
	userRole, _ := roleVal.(domain.Role)

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid listing id")
		return
	}

	var req createListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	listing := &domain.Listing{
		ID:          id,
		BrandID:     req.BrandID,
		ModelID:     req.ModelID,
		Year:        req.Year,
		Price:       req.Price,
		Mileage:     req.Mileage,
		VIN:         req.VIN,
		Description: req.Description,
	}

	updated, err := h.listingUsecase.UpdateListing(r.Context(), userID, userRole, listing)
	if err != nil {
		if errors.Is(err, domain.ErrListingNotFound) {
			response.Error(w, http.StatusNotFound, "listing not found")
			return
		}
		if errors.Is(err, domain.ErrListingForbidden) {
			response.Error(w, http.StatusForbidden, "access denied")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update listing")
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

// Delete godoc
// @Summary      Удаление объявления
// @Tags         listings
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "ID объявления"
// @Success      200 {object} response.Response{data=string}
// @Router       /listings/{id} [delete]
func (h *ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roleVal := r.Context().Value(middleware.UserRoleKey)
	userRole, _ := roleVal.(domain.Role)

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid listing id")
		return
	}

	if err := h.listingUsecase.DeleteListing(r.Context(), userID, userRole, id); err != nil {
		if errors.Is(err, domain.ErrListingNotFound) {
			response.Error(w, http.StatusNotFound, "listing not found")
			return
		}
		if errors.Is(err, domain.ErrListingForbidden) {
			response.Error(w, http.StatusForbidden, "access denied")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete listing")
		return
	}

	response.JSON(w, http.StatusOK, "listing successfully deleted")
}

// UploadImage godoc
// @Summary      Загрузка фотографии к объявлению
// @Tags         listings
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "ID объявления"
// @Param        image formData file true "Файл картинки (jpg, png, webp)"
// @Success      201 {object} response.Response{data=domain.ListingImage}
// @Router       /listings/{id}/images [post]
func (h *ListingHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roleVal := r.Context().Value(middleware.UserRoleKey)
	userRole, _ := roleVal.(domain.Role)

	idStr := chi.URLParam(r, "id")
	listingID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid listing id")
		return
	}

	// Ограничиваем считывание тела запроса максимум 10 МБ в память
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "missing image field in form-data")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to read uploaded file")
		return
	}

	img, err := h.listingUsecase.UploadImage(r.Context(), userID, userRole, listingID, header.Filename, fileBytes)
	if err != nil {
		if errors.Is(err, domain.ErrListingForbidden) {
			response.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if errors.Is(err, domain.ErrInvalidImageFormat) || errors.Is(err, domain.ErrImageTooLarge) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to upload image")
		return
	}

	response.JSON(w, http.StatusCreated, img)
}

// GetBrands godoc
// @Summary      Список всех марок авто
// @Tags         catalog
// @Produce      json
// @Success      200 {object} response.Response{data=[]domain.CarBrand}
// @Router       /brands [get]
func (h *ListingHandler) GetBrands(w http.ResponseWriter, r *http.Request) {
	brands, err := h.listingUsecase.GetBrands(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get brands")
		return
	}
	response.JSON(w, http.StatusOK, brands)
}

// GetModels godoc
// @Summary      Список моделей марки
// @Tags         catalog
// @Produce      json
// @Param        brand_id path int true "ID марки"
// @Success      200 {object} response.Response{data=[]domain.CarModel}
// @Router       /brands/{brand_id}/models [get]
func (h *ListingHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	brandIDStr := chi.URLParam(r, "brand_id")
	brandID, err := strconv.Atoi(brandIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid brand id")
		return
	}

	models, err := h.listingUsecase.GetModels(r.Context(), brandID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get models")
		return
	}
	response.JSON(w, http.StatusOK, models)
}
