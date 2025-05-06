package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go_postgres/internal/service"

	"go.uber.org/zap"
)

type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
}

func NewUserHandler(userService service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req service.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			h.respondWithError(w, http.StatusConflict, "user already exists")
		} else {
			h.logger.Error("failed to create user", zap.Error(err))
			h.respondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get user
	user, err := h.userService.GetUser(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			h.logger.Error("Failed to get user", zap.Error(err))
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	if pageStr != "" {
		pageVal, err := strconv.Atoi(pageStr)
		if err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	pageSize := 10
	if pageSizeStr != "" {
		pageSizeVal, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSizeVal > 0 && pageSizeVal <= 100 {
			pageSize = pageSizeVal
		}
	}

	// Get users
	users, total, err := h.userService.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create response with pagination info
	response := map[string]interface{}{
		"users":       users,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse request body
	var req service.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Update user
	user, err := h.userService.UpdateUser(r.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			h.logger.Error("Failed to update user", zap.Error(err))
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Delete user
	err = h.userService.DeleteUser(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			h.logger.Error("Failed to delete user", zap.Error(err))
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return success with no content
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Authenticate user
	user, err := h.userService.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		} else {
			h.logger.Error("Failed to authenticate user", zap.Error(err))
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// For a real application, you would generate a JWT token here
	// and return it in the response

	h.respondWithJSON(w, http.StatusOK, user)
}

// respondWithError sends an error response
func (h *UserHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func (h *UserHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	w.WriteHeader(code)

	// Encode payload to JSON
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}
