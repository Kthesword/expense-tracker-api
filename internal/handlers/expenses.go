package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kthesword/expense-tracker-api/internal/middleware"
	"github.com/kthesword/expense-tracker-api/internal/models"
	"gorm.io/gorm"
)

const dateLayout = "2006-01-02"

// ExpenseHandler handles CRUD on a user's expenses.
type ExpenseHandler struct {
	DB *gorm.DB
}

type createExpenseRequest struct {
	Title       string          `json:"title" binding:"required,max=200"`
	Description string          `json:"description" binding:"max=1000"`
	Amount      float64         `json:"amount" binding:"required,gt=0"`
	Category    models.Category `json:"category" binding:"required"`
	Date        string          `json:"date"` // YYYY-MM-DD, defaults to today
}

func userID(c *gin.Context) uint { return c.GetUint(middleware.UserIDKey) }

func today() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

// Create adds an expense for the authenticated user.
func (h *ExpenseHandler) Create(c *gin.Context) {
	var req createExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !req.Category.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category", "allowed": models.AllCategories})
		return
	}
	date := today()
	if req.Date != "" {
		d, err := time.Parse(dateLayout, req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
			return
		}
		date = d
	}

	exp := models.Expense{
		UserID:      userID(c),
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Amount:      req.Amount,
		Category:    req.Category,
		Date:        date,
	}
	if err := h.DB.Create(&exp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create expense"})
		return
	}
	c.JSON(http.StatusCreated, exp)
}
