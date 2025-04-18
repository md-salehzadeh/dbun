package model

import (
	"fmt"
	"strconv"
)

// TableMetadata contains metadata about a database table
type TableMetadata struct {
	Name    string
	Columns []ColumnMetadata
}

// ColumnMetadata contains metadata about a table column
type ColumnMetadata struct {
	Name     string
	Type     string
	Nullable bool
	Key      string
}

// RowData represents a generic row of data from any table
type RowData map[string]interface{}

// ViewMode represents the different viewing modes in the application
type ViewMode string

// Constants for view modes
const (
	DataMode      ViewMode = "Data"
	StructureMode ViewMode = "Structure"
	IndicesMode   ViewMode = "Indices"
)

// Sample data models for fallback data
type User struct {
	ID       int
	Username string
	Email    string
	Active   bool
}

type Order struct {
	ID         int
	UserID     int
	TotalPrice float64
	Status     string
}

type Product struct {
	ID       int
	Name     string
	Price    float64
	Category string
}

type Category struct {
	ID   int
	Name string
	Slug string
}

// GenerateSampleData creates sample data for testing without a database
func GenerateSampleData() ([]User, []Order, []Product, []Category) {
	users := []User{
		{ID: 1, Username: "johndoe", Email: "john@example.com", Active: true},
		{ID: 2, Username: "janedoe", Email: "jane@example.com", Active: true},
		{ID: 3, Username: "bobsmith", Email: "bob@example.com", Active: false},
		{ID: 4, Username: "alicejones", Email: "alice@example.com", Active: true},
		{ID: 5, Username: "mikebrown", Email: "mike@example.com", Active: true},
	}

	orders := []Order{
		{ID: 101, UserID: 1, TotalPrice: 125.99, Status: "Completed"},
		{ID: 102, UserID: 2, TotalPrice: 89.50, Status: "Processing"},
		{ID: 103, UserID: 1, TotalPrice: 45.75, Status: "Shipped"},
		{ID: 104, UserID: 3, TotalPrice: 210.25, Status: "Pending"},
		{ID: 105, UserID: 4, TotalPrice: 55.00, Status: "Completed"},
	}

	products := []Product{
		{ID: 201, Name: "Laptop", Price: 999.99, Category: "Electronics"},
		{ID: 202, Name: "Headphones", Price: 129.99, Category: "Electronics"},
		{ID: 203, Name: "Coffee Maker", Price: 79.50, Category: "Appliances"},
		{ID: 204, Name: "Running Shoes", Price: 89.95, Category: "Footwear"},
		{ID: 205, Name: "Desk Chair", Price: 199.99, Category: "Furniture"},
	}

	categories := []Category{
		{ID: 301, Name: "Electronics", Slug: "electronics"},
		{ID: 302, Name: "Appliances", Slug: "appliances"},
		{ID: 303, Name: "Footwear", Slug: "footwear"},
		{ID: 304, Name: "Furniture", Slug: "furniture"},
		{ID: 305, Name: "Books", Slug: "books"},
	}

	return users, orders, products, categories
}

// Helper functions for value parsing
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// Truncates text with ellipsis if it exceeds width
func TruncateWithEllipsis(text string, width int) string {
	if len(text) <= width {
		return text
	}
	
	if width <= 3 {
		return text[:width]
	}
	
	return text[:width-3] + "..."
}

// FormatValue converts an interface{} value to a string for display
func FormatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		// Consider using a specific precision if needed, e.g., "%.2f"
		return fmt.Sprintf("%g", v) // %g is often good for floats
	case []byte:
		// Assume byte slices are strings (common in database/sql)
		return string(v)
	case string:
		return v
	// Add other types as needed, e.g., time.Time
	// case time.Time:
	//  return v.Format("2006-01-02 15:04:05")
	default:
		// Fallback for other types
		return fmt.Sprintf("%v", v)
	}
}