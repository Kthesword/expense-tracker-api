// Package handlers contains the HTTP handlers.
package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kthesword/expense-tracker-api/internal/auth"
	"github.com/kthesword/expense-tracker-api/internal/models"
	"gorm.io/gorm"
)

// AuthHandler handles authentication.
type AuthHandler struct {
	DB     *gorm.DB
	Tokens *auth.TokenManager
}

type signupRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"` // bcrypt max 72 bytes
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Signup creates a new user and returns a JWT.
func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process password"})
		return
	}

	user := models.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: hash,
	}
	if err := h.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	token, err := h.Tokens.Generate(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user, "token": token})
}

// Login verifies credentials and returns a JWT.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&user).Error
	if err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := h.Tokens.Generate(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}
