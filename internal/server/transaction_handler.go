package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yusuf4ktas/backend-project/internal/service"
	"github.com/yusuf4ktas/backend-project/internal/worker"
)

type TransactionHandler struct {
	dispatcher *worker.Dispatcher
	service    service.TransactionService
}

func NewTransactionHandler(d *worker.Dispatcher, s service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		dispatcher: d,
		service:    s,
	}
}

// The request struct lacks from "from_id" field because the sender's ID comes from the token.
type transferRequest struct {
	ToUserID int64   `json:"to_user_id"`
	Amount   float64 `json:"amount"`
}

type creditRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type debitRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) *apiError {
	// Get the authenticated user's ID from the context.
	fromUserID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		return &apiError{Status: http.StatusInternalServerError, Message: "User ID not found in context"}
	}

	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return &apiError{Status: http.StatusBadRequest, Message: "Invalid request body"}
	}

	// Creation of job from the request
	job := worker.Job{
		FromUserID:      fromUserID,
		ToUserID:        req.ToUserID,
		Amount:          req.Amount,
		TransactionType: "transfer",
	}

	// Adding job to the dispatcher queue
	h.dispatcher.AddJob(job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted
	json.NewEncoder(w).Encode(map[string]string{"message": "Transaction queued for processing."})

	return nil
}

func (h *TransactionHandler) Credit(w http.ResponseWriter, r *http.Request) *apiError {
	var req creditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return &apiError{Status: http.StatusBadRequest, Message: "Invalid request body"}
	}

	job := worker.Job{
		FromUserID:      0,
		ToUserID:        req.UserID,
		Amount:          req.Amount,
		TransactionType: "credit",
	}
	h.dispatcher.AddJob(job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Credit transaction queued."})
	return nil
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) *apiError {
	var req debitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return &apiError{Status: http.StatusBadRequest, Message: "Invalid request body"}
	}

	job := worker.Job{
		FromUserID:      req.UserID,
		ToUserID:        0,
		Amount:          req.Amount,
		TransactionType: "debit",
	}
	h.dispatcher.AddJob(job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Debit transaction queued."})
	return nil
}

func (h *TransactionHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) *apiError {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		return &apiError{Status: http.StatusInternalServerError, Message: "User ID not found in context"}
	}

	transactions, err := h.service.GetTransactionHistory(r.Context(), userID) // Assume GetHistory exists on service
	if err != nil {
		return &apiError{Status: http.StatusInternalServerError, Message: "Failed to retrieve transaction history"}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
	return nil
}

func (h *TransactionHandler) GetByTransactionID(w http.ResponseWriter, r *http.Request) *apiError {
	idStr := chi.URLParam(r, "id")
	transactionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &apiError{Status: http.StatusBadRequest, Message: "Invalid transaction ID format"}
	}

	transaction, err := h.service.GetByTransactionID(r.Context(), transactionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &apiError{Status: http.StatusNotFound, Message: "Transaction not found"}
		}
		return &apiError{Status: http.StatusInternalServerError, Message: "Failed to retrieve transaction"}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transaction)
	return nil
}
