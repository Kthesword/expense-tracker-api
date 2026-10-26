package handlers

import (
	"errors"
	"net/http"
	"strconv"
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

type updateExpenseRequest struct {
	Title       *string          `json:"title" binding:"omitempty,max=200"`
	Description *string          `json:"description" binding:"omitempty,max=1000"`
	Amount      *float64         `json:"amount" binding:"omitempty,gt=0"`
	Category    *models.Category `json:"category"`
	Date        *string          `json:"date"`
}

func userID(c *gin.Context) uint { return c.GetUint(middleware.UserIDKey) }

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
		return 0, false
	}
	return uint(id), true
}

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

// List returns the user's expenses, optionally filtered by period.
//
//	?filter=week|month|3months|custom   (custom needs ?start=YYYY-MM-DD&end=YYYY-MM-DD)
//	?category=Food
func (h *ExpenseHandler) List(c *gin.Context) {
	q := h.DB.Where("user_id = ?", userID(c))
	now := today()

	switch f := c.Query("filter"); f {
	case "":
		// no date filter
	case "week":
		q = q.Where("date >= ?", now.AddDate(0, 0, -7))
	case "month":
		q = q.Where("date >= ?", now.AddDate(0, -1, 0))
	case "3months":
		q = q.Where("date >= ?", now.AddDate(0, -3, 0))
	case "custom":
		start, err1 := time.Parse(dateLayout, c.Query("start"))
		end, err2 := time.Parse(dateLayout, c.Query("end"))
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "custom filter requires start and end as YYYY-MM-DD"})
			return
		}
		if end.Before(start) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end must not be before start"})
			return
		}
		q = q.Where("date BETWEEN ? AND ?", start, end)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "filter must be one of: week, month, 3months, custom"})
		return
	}

	if cat := c.Query("category"); cat != "" {
		if !models.Category(cat).Valid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category", "allowed": models.AllCategories})
			return
		}
		q = q.Where("category = ?", cat)
	}

	expenses := []models.Expense{}
	if err := q.Order("date DESC, id DESC").Find(&expenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch expenses"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": len(expenses), "expenses": expenses})
}

// Get returns one expense owned by the user.
func (h *ExpenseHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var exp models.Expense
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).First(&exp).Error; err != nil {
		h.notFoundOrErr(c, err)
		return
	}
	c.JSON(http.StatusOK, exp)
}

// Update partially updates an expense owned by the user.
func (h *ExpenseHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req updateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exp models.Expense
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).First(&exp).Error; err != nil {
		h.notFoundOrErr(c, err)
		return
	}

	if req.Title != nil {
		exp.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		exp.Description = *req.Description
	}
	if req.Amount != nil {
		exp.Amount = *req.Amount
	}
	if req.Category != nil {
		if !req.Category.Valid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category", "allowed": models.AllCategories})
			return
		}
		exp.Category = *req.Category
	}
	if req.Date != nil {
		d, err := time.Parse(dateLayout, *req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
			return
		}
		exp.Date = d
	}

	if err := h.DB.Save(&exp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update expense"})
		return
	}
	c.JSON(http.StatusOK, exp)
}

// Delete removes an expense owned by the user.
func (h *ExpenseHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).Delete(&models.Expense{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete expense"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ExpenseHandler) notFoundOrErr(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
}
