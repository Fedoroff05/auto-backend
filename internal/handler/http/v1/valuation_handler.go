package v1

import (
	"net/http"
	"strconv"

	"github.com/Fedoroff05/auto-backend/internal/handler/http/response"
	"github.com/Fedoroff05/auto-backend/internal/usecase"
)

type ValuationHandler struct {
	valuationUsecase *usecase.ValuationUsecase
}

func NewValuationHandler(valuationUsecase *usecase.ValuationUsecase) *ValuationHandler {
	return &ValuationHandler{valuationUsecase: valuationUsecase}
}

// EstimatePrice godoc
// @Summary      Оценка рыночной стоимости автомобиля
// @Tags         valuation
// @Produce      json
// @Param        brand_id query int true "ID марки"
// @Param        model_id query int true "ID модели"
// @Param        year query int true "Год выпуска"
// @Success      200 {object} response.Response{data=domain.MarketValuationResult}
// @Router       /valuation/estimate [get]
func (h *ValuationHandler) EstimatePrice(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	brandID, err1 := strconv.Atoi(q.Get("brand_id"))
	modelID, err2 := strconv.Atoi(q.Get("model_id"))
	year, err3 := strconv.Atoi(q.Get("year"))

	if err1 != nil || err2 != nil || err3 != nil {
		response.Error(w, http.StatusBadRequest, "brand_id, model_id and year are required query params")
		return
	}

	result, err := h.valuationUsecase.EstimatePrice(r.Context(), brandID, modelID, year)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}
