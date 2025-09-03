package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yusuf4ktas/backend-project/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) *apiError {
	var req registerRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return &apiError{Status: http.StatusBadRequest, Message: "Invalid request body"}
	}

	createdUser, err := h.userService.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return &apiError{Status: http.StatusInternalServerError, Message: err.Error()}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdUser)

	return nil
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) *apiError {
	// Get the 'id' from the URL path
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &apiError{Status: http.StatusBadRequest, Message: "Invalid user ID format"}
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &apiError{Status: http.StatusNotFound, Message: "User not found"}
		}
		return &apiError{Status: http.StatusInternalServerError, Message: err.Error()}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)

	return nil
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) *apiError {
	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		return &apiError{Status: http.StatusInternalServerError, Message: err.Error()}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("ERROR: failed to write GetAll users response: %v", err)
	}
	return nil
}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) *apiError {
	idStr := chi.URLParam(r, "id")
	userIDToDelete, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &apiError{Status: http.StatusBadRequest, Message: "Invalid user ID specified"}
	}

	// Get the ID and details of the user making the request from the context.
	requestingUserID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		return &apiError{Status: http.StatusInternalServerError, Message: "User ID not found in context"}
	}
	requestingUser, err := h.userService.GetByID(r.Context(), requestingUserID)
	if err != nil {
		return &apiError{Status: http.StatusInternalServerError, Message: "Could not retrieve requesting user's details"}
	}

	// Allow if the user is an admin OR if the user is deleting their own account.
	if requestingUser.Role != "admin" && requestingUserID != userIDToDelete {
		return &apiError{Status: http.StatusForbidden, Message: "You do not have permission to delete this user"}
	}

	if err := h.userService.Delete(r.Context(), userIDToDelete); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &apiError{Status: http.StatusNotFound, Message: "User to delete not found"}
		}
		return &apiError{Status: http.StatusInternalServerError, Message: "Failed to delete user"}
	}

	w.WriteHeader(http.StatusNoContent) //successful deletion
	return nil
}
