// Package models defines the database entities.
package models

import "time"

// Category is a fixed set of expense categories.
type Category string

const (
	Food        Category = "Food"
	Leisure     Category = "Leisure"
	Electronics Category = "Electronics"
	Utilities   Category = "Utilities"
	Clothing    Category = "Clothing"
	Health      Category = "Health"
)

// AllCategories lists every valid category.
var AllCategories = []Category{Food, Leisure, Electronics, Utilities, Clothing, Health}

// Valid reports whether c is one of the allowed categories.
func (c Category) Valid() bool {
	for _, v := range AllCategories {
		if v == c {
			return true
		}
	}
	return false
}

// User is an account owner.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"` // never serialised
	CreatedAt    time.Time `json:"created_at"`
}

// Expense belongs to exactly one user.
type Expense struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index:idx_expenses_user_date,priority:1" json:"-"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Amount      float64   `gorm:"type:numeric(12,2);not null" json:"amount"`
	Category    Category  `gorm:"type:varchar(20);not null" json:"category"`
	Date        time.Time `gorm:"type:date;not null;index:idx_expenses_user_date,priority:2" json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
