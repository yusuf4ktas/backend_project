package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yusuf4ktas/backend-project/internal/service"
)

type BalanceHandler struct {
	balanceService service.BalanceService
}

func NewBalanceHandler(s service.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: s}
}

func (h *BalanceHandler) GetCurrentBalance(w http.ResponseWriter, r *http.Request) *apiError {
	// Get the authenticated user's ID from the context.
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		return &apiError{Status: http.StatusInternalServerError, Message: "User ID not found in context"}
	}

	//Call the service to get the current balance.
	balance, err := h.balanceService.GetCurrent(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &apiError{Status: http.StatusNotFound, Message: "Balance for user not found"}
		}
		return &apiError{Status: http.StatusInternalServerError, Message: "Failed to retrieve balance"}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
	return nil
}
