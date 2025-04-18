package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Sample data structures for tables
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

type model struct {
	tables         []string // List of tables
	selectedIdx    int      // Currently selected table index
	activeTableIdx int      // Currently active table index
	mode           string   // Mode: "Data", "Structure", "Indices"
	width          int      // Terminal width
	height         int      // Terminal height
	focusLeft      bool     // Focus on left (true) or right (false) box
	
	// Sample data for each table
	users      []User
	orders     []Order
	products   []Product
	categories []Category
}

// Initialize sample data for tables
func initSampleData() ([]User, []Order, []Product, []Category) {
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

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit // Exit app

		// Switch focus between left and right boxes
		case "1", "h":
			m.focusLeft = true

		case "2", "l":
			m.focusLeft = false

		case "enter":
			if m.focusLeft {
				// Activate the selected table
				m.activeTableIdx = m.selectedIdx
			}

		default:
			if m.focusLeft {
				// Navigation within the left box (tables list)
				switch msg.String() {
				case "up", "k":
					if m.selectedIdx > 0 {
						m.selectedIdx--
					}
				case "down", "j":
					if m.selectedIdx < len(m.tables)-1 {
						m.selectedIdx++
					}
				}
			} else {
				// Navigation or actions in the right box can be added here
				// For now, we can leave it empty or handle right box specific keys
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	// Ensure we have valid dimensions
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	doc := strings.Builder{}

	// Calculate dynamic widths and heights based on terminal size
	sidebarWidth := int(0.2 * float64(m.width)) // 20% of terminal width
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}

	// Calculate height for main content area
	buttonHeightWithPadding := 3              // Height for the tab bar
	statusBarHeight := 1                      // Height for status bar
	mainHeight := m.height - buttonHeightWithPadding - statusBarHeight - 2 // Account for borders

	// Styles for active and inactive borders
	activeBorderColor := lipgloss.Color("#FF00FF")   // Magenta for active box
	inactiveBorderColor := lipgloss.Color("#444444") // Grey for inactive box

	// Sidebar Style
	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(mainHeight).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Left)

	// Main Box Style
	mainBoxWidth := m.width - sidebarWidth - 4 // Account for borders
	if mainBoxWidth < 20 {
		mainBoxWidth = 20
	}
	mainBoxStyle := lipgloss.NewStyle().
		Width(mainBoxWidth).
		Height(mainHeight).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Left)

	// Set border colors based on focus
	if m.focusLeft {
		sidebarStyle = sidebarStyle.
			BorderForeground(activeBorderColor)
		mainBoxStyle = mainBoxStyle.
			BorderForeground(inactiveBorderColor)
	} else {
		sidebarStyle = sidebarStyle.
			BorderForeground(inactiveBorderColor)
		mainBoxStyle = mainBoxStyle.
			BorderForeground(activeBorderColor)
	}

	// Button Styles for modes
	buttonWidth := int(float64(m.width-6) / 3.0) // Divide space among three buttons, with some extra space
	activeButtonBorderColor := lipgloss.Color("#FF00FF")
	inactiveButtonBorderColor := lipgloss.Color("#444444")
	
	// Base button style - use tabs style instead of bordered boxes
	tabStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Bold(true).
		Width(buttonWidth)

	// Prepare button styles based on current mode
	dataButtonStyle := tabStyle.Copy()
	structureButtonStyle := tabStyle.Copy()
	indicesButtonStyle := tabStyle.Copy()
	
	// Set active styling based on current mode
	if m.mode == "Data" {
		dataButtonStyle = dataButtonStyle.
			Foreground(activeButtonBorderColor).
			Underline(true)
		structureButtonStyle = structureButtonStyle.
			Foreground(inactiveButtonBorderColor)
		indicesButtonStyle = indicesButtonStyle.
			Foreground(inactiveButtonBorderColor)
	} else if m.mode == "Structure" {
		dataButtonStyle = dataButtonStyle.
			Foreground(inactiveButtonBorderColor)
		structureButtonStyle = structureButtonStyle.
			Foreground(activeButtonBorderColor).
			Underline(true)
		indicesButtonStyle = indicesButtonStyle.
			Foreground(inactiveButtonBorderColor)
	} else {
		dataButtonStyle = dataButtonStyle.
			Foreground(inactiveButtonBorderColor)
		structureButtonStyle = structureButtonStyle.
			Foreground(inactiveButtonBorderColor)
		indicesButtonStyle = indicesButtonStyle.
			Foreground(activeButtonBorderColor).
			Underline(true)
	}

	// Status Bar Style
	statusBarStyle := lipgloss.NewStyle().
		Width(m.width).
		Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
		Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#FF5F87")).
		Padding(0, 1).
		MarginRight(1)

	encodingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#A550DF")).
		Padding(0, 1)

	fishCakeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#6124DF")).
		Padding(0, 1)

	statusTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
		Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	// Sidebar: Table List
	var sidebar strings.Builder

	// Styles for selected and normal items
	selectedItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF00FF"))
	activeItemStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#444444")).
		Foreground(lipgloss.Color("#FFFFFF"))
	normalItemStyle := lipgloss.NewStyle()

	for i, table := range m.tables {
		var line string
		cursor := " " // Default cursor

		if m.selectedIdx == i && m.activeTableIdx == i {
			cursor = ">"
			line = activeItemStyle.Render(fmt.Sprintf("%s %s", cursor, table))
		} else if m.selectedIdx == i {
			cursor = ">"
			line = selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, table))
		} else if m.activeTableIdx == i {
			line = activeItemStyle.Render(fmt.Sprintf("%s %s", cursor, table))
		} else {
			line = normalItemStyle.Render(fmt.Sprintf("%s %s", cursor, table))
		}

		sidebar.WriteString(line + "\n")
	}

	sidebarView := sidebarStyle.Render(sidebar.String())

	// Main Box: Show selected table data based on mode and selected table
	var mainContent string
	activeTable := m.tables[m.activeTableIdx]
	
	if m.mode == "Data" {
		// Display table data based on the selected table
		switch activeTable {
		case "users":
			mainContent = m.renderUsersTable()
		case "orders":
			mainContent = m.renderOrdersTable()
		case "products":
			mainContent = m.renderProductsTable()
		case "categories":
			mainContent = m.renderCategoriesTable()
		default:
			mainContent = "Unknown table selected"
		}
	} else if m.mode == "Structure" {
		// Display table structure based on the selected table
		switch activeTable {
		case "users":
			mainContent = "User Structure:\n\nID: int\nUsername: string\nEmail: string\nActive: bool"
		case "orders":
			mainContent = "Order Structure:\n\nID: int\nUserID: int\nTotalPrice: float64\nStatus: string"
		case "products":
			mainContent = "Product Structure:\n\nID: int\nName: string\nPrice: float64\nCategory: string"
		case "categories":
			mainContent = "Category Structure:\n\nID: int\nName: string\nSlug: string"
		default:
			mainContent = "Unknown table selected"
		}
	} else if m.mode == "Indices" {
		// Display indices information based on the selected table
		switch activeTable {
		case "users":
			mainContent = "User Indices:\n\nPrimary Key: ID\nIndex: Username (unique)\nIndex: Email (unique)"
		case "orders":
			mainContent = "Order Indices:\n\nPrimary Key: ID\nIndex: UserID"
		case "products":
			mainContent = "Product Indices:\n\nPrimary Key: ID\nIndex: Category"
		case "categories":
			mainContent = "Category Indices:\n\nPrimary Key: ID\nIndex: Slug (unique)"
		default:
			mainContent = "Unknown table selected"
		}
	} else {
		mainContent = fmt.Sprintf("Unknown mode: %s", m.mode)
	}
	
	mainBoxView := mainBoxStyle.Render(mainContent)

	// Top Buttons for switching views - styled as tabs
	buttonBar := lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 0, 0, 2).  // Top, Right, Bottom, Left padding
		Background(lipgloss.Color("#222222"))

	buttons := lipgloss.JoinHorizontal(lipgloss.Top,
		dataButtonStyle.Render("Data"),
		structureButtonStyle.Render("Structure"),
		indicesButtonStyle.Render("Indices"),
	)
	
	buttonSection := buttonBar.Render(buttons)
	
	// Combine the main views
	layout := lipgloss.JoinVertical(lipgloss.Top, 
		buttonSection,
		lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainBoxView),
	)

	// Write the layout to the document
	doc.WriteString(layout)

	// Status bar
	{
		w := lipgloss.Width

		statusKey := statusStyle.Render("STATUS")
		encoding := encodingStyle.Render("UTF-8")
		fishCake := fishCakeStyle.Render("🍥 Fish Cake")
		statusVal := statusTextStyle.
			Width(m.width - w(statusKey) - w(encoding) - w(fishCake) - 5).
			Render("Ravishing")

		bar := lipgloss.JoinHorizontal(lipgloss.Top,
			statusKey,
			statusVal,
			encoding,
			fishCake,
		)

		// Append status bar to the document
		doc.WriteString("\n" + statusBarStyle.Render(bar))
	}

	// Return the complete document
	return doc.String()
}

// Helper methods to render table data
func (m model) renderUsersTable() string {
	var content strings.Builder
	
	// Define column widths for consistent alignment
	colWidths := []int{5, 20, 30, 10}
	
	// Define table styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#222222")).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	
	rowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))
	
	// Add row numbers column at the beginning
	content.WriteString(fmt.Sprintf("%-4s ", "#"))
	
	// Format the headers
	headers := []string{"ID", "USERNAME", "EMAIL", "ACTIVE"}
	for i, header := range headers {
		content.WriteString(headerStyle.Render(fmt.Sprintf("%-*s ", colWidths[i], header)))
	}
	content.WriteString("\n")
	
	// Table rows with alternating styles
	for i, user := range m.users {
		// Row number
		content.WriteString(fmt.Sprintf("%-4d ", i+1))
		
		// Row content with cell styling
		activeStr := "No"
		if user.Active {
			activeStr = "Yes"
		}
		
		// Style based on even/odd row
		var rowStyleToUse lipgloss.Style
		if i%2 == 0 {
			rowStyleToUse = rowStyle
		} else {
			rowStyleToUse = cellStyle
		}
		
		// Format each cell
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*d ", colWidths[0], user.ID)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*s ", colWidths[1], user.Username)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*s ", colWidths[2], user.Email)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*s ", colWidths[3], activeStr)))
		
		content.WriteString("\n")
	}
	
	return content.String()
}

func (m model) renderOrdersTable() string {
	var content strings.Builder
	
	// Define column widths for consistent alignment
	colWidths := []int{5, 10, 15, 20}
	
	// Define table styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#222222")).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	
	rowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))
	
	// Add row numbers column at the beginning
	content.WriteString(fmt.Sprintf("%-4s ", "#"))
	
	// Format the headers
	headers := []string{"ID", "USER ID", "TOTAL PRICE", "STATUS"}
	for i, header := range headers {
		content.WriteString(headerStyle.Render(fmt.Sprintf("%-*s ", colWidths[i], header)))
	}
	content.WriteString("\n")
	
	// Table rows with alternating styles
	for i, order := range m.orders {
		// Row number
		content.WriteString(fmt.Sprintf("%-4d ", i+1))
		
		// Style based on even/odd row
		var rowStyleToUse lipgloss.Style
		if i%2 == 0 {
			rowStyleToUse = rowStyle
		} else {
			rowStyleToUse = cellStyle
		}
		
		// Format each cell
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*d ", colWidths[0], order.ID)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*d ", colWidths[1], order.UserID)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("$%-*.2f ", colWidths[2]-1, order.TotalPrice)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*s ", colWidths[3], order.Status)))
		
		content.WriteString("\n")
	}
	
	return content.String()
}

func (m model) renderProductsTable() string {
	var content strings.Builder
	
	// Define column widths for consistent alignment
	colWidths := []int{5, 25, 15, 20}
	
	// Define table styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#222222")).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	
	rowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))
	
	// Add row numbers column at the beginning
	content.WriteString(fmt.Sprintf("%-4s ", "#"))
	
	// Format the headers
	headers := []string{"ID", "NAME", "PRICE", "CATEGORY"}
	for i, header := range headers {
		content.WriteString(headerStyle.Render(fmt.Sprintf("%-*s ", colWidths[i], header)))
	}
	content.WriteString("\n")
	
	// Table rows with alternating styles
	for i, product := range m.products {
		// Row number
		content.WriteString(fmt.Sprintf("%-4d ", i+1))
		
		// Style based on even/odd row
		var rowStyleToUse lipgloss.Style
		if i%2 == 0 {
			rowStyleToUse = rowStyle
		} else {
			rowStyleToUse = cellStyle
		}
		
		// Format each cell
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*d ", colWidths[0], product.ID)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*s ", colWidths[1], product.Name)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("$%-*.2f ", colWidths[2]-1, product.Price)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*s ", colWidths[3], product.Category)))
		
		content.WriteString("\n")
	}
	
	return content.String()
}

func (m model) renderCategoriesTable() string {
	var content strings.Builder
	
	// Define column widths for consistent alignment
	colWidths := []int{5, 25, 25}
	
	// Define table styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#222222")).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	
	rowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))
	
	// Add row numbers column at the beginning
	content.WriteString(fmt.Sprintf("%-4s ", "#"))
	
	// Format the headers
	headers := []string{"ID", "NAME", "SLUG"}
	for i, header := range headers {
		content.WriteString(headerStyle.Render(fmt.Sprintf("%-*s ", colWidths[i], header)))
	}
	content.WriteString("\n")
	
	// Table rows with alternating styles
	for i, category := range m.categories {
		// Row number
		content.WriteString(fmt.Sprintf("%-4d ", i+1))
		
		// Style based on even/odd row
		var rowStyleToUse lipgloss.Style
		if i%2 == 0 {
			rowStyleToUse = rowStyle
		} else {
			rowStyleToUse = cellStyle
		}
		
		// Format each cell
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*d ", colWidths[0], category.ID)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*s ", colWidths[1], category.Name)))
		content.WriteString(rowStyleToUse.Render(fmt.Sprintf("%-*s ", colWidths[2], category.Slug)))
		
		content.WriteString("\n")
	}
	
	return content.String()
}

func main() {
	users, orders, products, categories := initSampleData()
	m := model{
		tables:         []string{"users", "orders", "products", "categories"},
		selectedIdx:    0,
		activeTableIdx: 0, // Initialize active table index
		mode:           "Data", // Start with Data view
		width:          80,     // Default width
		height:         24,     // Default height
		focusLeft:      true,   // Start with focus on left box
		users:          users,
		orders:         orders,
		products:       products,
		categories:     categories,
	}

	p := tea.NewProgram(&m, tea.WithAltScreen()) // Fullscreen app

	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
