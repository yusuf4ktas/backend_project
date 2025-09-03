package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yusuf4ktas/backend-project/internal/service"
)

type AuthHandler struct {
	userService service.UserService
	jwtSecret   []byte
}

func NewAuthHandler(us service.UserService, secret []byte) *AuthHandler {
	return &AuthHandler{
		userService: us,
		jwtSecret:   secret,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) *apiError {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return &apiError{Status: http.StatusBadRequest, Message: "Invalid request body"}
	}

	user, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		return &apiError{Status: http.StatusUnauthorized, Message: "Invalid email or password"}
	}

	// JWT generation
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 72).Unix(), //Token expire time is set to 72 hours, can be adjusted
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return &apiError{Status: http.StatusInternalServerError, Message: "Failed to generate token"}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
	return nil
}
