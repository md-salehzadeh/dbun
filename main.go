package main

import (
	"fmt"
	"os"
	"strconv"
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
	
	// Editing state
	editing       bool   // Whether we're in editing mode
	cursorRow     int    // Current cursor row in the table
	cursorCol     int    // Current cursor column in the table
	editBuffer    string // Buffer for the current edit
	editingField  string // Name of the field being edited
	showEditHelp  bool   // Whether to show editing help
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
		// Handle global keys first
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit // Exit app

		case "tab":
			// Toggle focus between left and right boxes
			m.focusLeft = !m.focusLeft
			// Reset editing state when switching focus
			if m.focusLeft && m.editing {
				m.editing = false
				m.editBuffer = ""
			}
			return m, nil

		// Switch focus between left and right boxes
		case "1":
			m.focusLeft = true
			// Reset editing state when switching focus
			if m.editing {
				m.editing = false
				m.editBuffer = ""
			}
			return m, nil

		case "2":
			m.focusLeft = false
			return m, nil

		case "?":
			// Toggle help display
			m.showEditHelp = !m.showEditHelp
			return m, nil
		}

		// Handle edit mode keys
		if m.editing {
			switch msg.String() {
			case "esc":
				// Cancel edit
				m.editing = false
				m.editBuffer = ""
				return m, nil

			case "enter":
				// Submit edit
				m.applyEdit()
				m.editing = false
				m.editBuffer = ""
				return m, nil

			case "backspace":
				// Delete last character
				if len(m.editBuffer) > 0 {
					m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
				}
				return m, nil

			default:
				// Only add printable characters (ASCII 32-126) to the edit buffer
				if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
					m.editBuffer += msg.String()
				}
				return m, nil
			}
		}

		// Handle normal mode keys
		if m.focusLeft {
			// Navigation within the left box (tables list)
			switch msg.String() {
			case "up", "k":
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
				return m, nil
			case "down", "j":
				if m.selectedIdx < len(m.tables)-1 {
					m.selectedIdx++
				}
				return m, nil
			case "enter":
				// Activate the selected table
				m.activeTableIdx = m.selectedIdx
				return m, nil
			}
		} else {
			// Navigation or actions in the right box (data view)
			if m.mode == "Data" {
				switch msg.String() {
				case "up", "k":
					if m.cursorRow > 0 {
						m.cursorRow--
					}
					return m, nil
				case "down", "j":
					// Max rows depends on the current table
					var maxRows int
					switch m.tables[m.activeTableIdx] {
					case "users":
						maxRows = len(m.users)
					case "orders":
						maxRows = len(m.orders)
					case "products":
						maxRows = len(m.products)
					case "categories":
						maxRows = len(m.categories)
					default:
						maxRows = 0
					}
					
					if m.cursorRow < maxRows-1 {
						m.cursorRow++
					}
					return m, nil
				case "left", "h":
					if m.cursorCol > 0 {
						m.cursorCol--
					}
					return m, nil
				case "right", "l":
					// Max columns depends on the current table
					var maxCols int
					switch m.tables[m.activeTableIdx] {
					case "users":
						maxCols = 4 // ID, Username, Email, Active
					case "orders":
						maxCols = 4 // ID, UserID, TotalPrice, Status
					case "products":
						maxCols = 4 // ID, Name, Price, Category
					case "categories":
						maxCols = 3 // ID, Name, Slug
					default:
						maxCols = 0
					}
					
					if m.cursorCol < maxCols-1 {
						m.cursorCol++
					}
					return m, nil
				case "enter", "e":
					// Enter edit mode for the current cell
					return m.enterEditMode(), nil
				}
			}
		}
	}

	return m, nil
}

// Enters edit mode for the current cell
func (m *model) enterEditMode() tea.Model {
	// Only allow editing in data mode
	if m.mode != "Data" {
		return m
	}
	
	// Set editing flag and prepare edit buffer
	m.editing = true
	m.editBuffer = m.getCurrentCellValue()
	m.editingField = m.getCurrentFieldName()
	
	return m
}

// Gets the name of the current field being edited
func (m *model) getCurrentFieldName() string {
	table := m.tables[m.activeTableIdx]
	
	switch table {
	case "users":
		fields := []string{"ID", "Username", "Email", "Active"}
		if m.cursorCol < len(fields) {
			return fields[m.cursorCol]
		}
	case "orders":
		fields := []string{"ID", "UserID", "TotalPrice", "Status"}
		if m.cursorCol < len(fields) {
			return fields[m.cursorCol]
		}
	case "products":
		fields := []string{"ID", "Name", "Price", "Category"}
		if m.cursorCol < len(fields) {
			return fields[m.cursorCol]
		}
	case "categories":
		fields := []string{"ID", "Name", "Slug"}
		if m.cursorCol < len(fields) {
			return fields[m.cursorCol]
		}
	}
	
	return ""
}

// Gets the current value of the selected cell
func (m *model) getCurrentCellValue() string {
	table := m.tables[m.activeTableIdx]
	
	switch table {
	case "users":
		if m.cursorRow >= 0 && m.cursorRow < len(m.users) {
			user := m.users[m.cursorRow]
			switch m.cursorCol {
			case 0:
				return fmt.Sprintf("%d", user.ID)
			case 1:
				return user.Username
			case 2:
				return user.Email
			case 3:
				if user.Active {
					return "Yes"
				}
				return "No"
			}
		}
	case "orders":
		if m.cursorRow >= 0 && m.cursorRow < len(m.orders) {
			order := m.orders[m.cursorRow]
			switch m.cursorCol {
			case 0:
				return fmt.Sprintf("%d", order.ID)
			case 1:
				return fmt.Sprintf("%d", order.UserID)
			case 2:
				return fmt.Sprintf("%.2f", order.TotalPrice)
			case 3:
				return order.Status
			}
		}
	case "products":
		if m.cursorRow >= 0 && m.cursorRow < len(m.products) {
			product := m.products[m.cursorRow]
			switch m.cursorCol {
			case 0:
				return fmt.Sprintf("%d", product.ID)
			case 1:
				return product.Name
			case 2:
				return fmt.Sprintf("%.2f", product.Price)
			case 3:
				return product.Category
			}
		}
	case "categories":
		if m.cursorRow >= 0 && m.cursorRow < len(m.categories) {
			category := m.categories[m.cursorRow]
			switch m.cursorCol {
			case 0:
				return fmt.Sprintf("%d", category.ID)
			case 1:
				return category.Name
			case 2:
				return category.Slug
			}
		}
	}
	
	return ""
}

// Apply the edit to the current cell
func (m *model) applyEdit() {
	table := m.tables[m.activeTableIdx]
	
	switch table {
	case "users":
		if m.cursorRow >= 0 && m.cursorRow < len(m.users) {
			user := &m.users[m.cursorRow]
			switch m.cursorCol {
			case 0:
				// ID: Parse int
				if id, err := parseInt(m.editBuffer); err == nil {
					user.ID = id
				}
			case 1:
				// Username: String
				user.Username = m.editBuffer
			case 2:
				// Email: String
				user.Email = m.editBuffer
			case 3:
				// Active: Bool (Yes/No)
				user.Active = m.editBuffer == "Yes" || m.editBuffer == "yes" || 
				             m.editBuffer == "true" || m.editBuffer == "TRUE" || 
				             m.editBuffer == "1"
			}
		}
	case "orders":
		if m.cursorRow >= 0 && m.cursorRow < len(m.orders) {
			order := &m.orders[m.cursorRow]
			switch m.cursorCol {
			case 0:
				// ID: Parse int
				if id, err := parseInt(m.editBuffer); err == nil {
					order.ID = id
				}
			case 1:
				// UserID: Parse int
				if userID, err := parseInt(m.editBuffer); err == nil {
					order.UserID = userID
				}
			case 2:
				// TotalPrice: Parse float
				if price, err := parseFloat(m.editBuffer); err == nil {
					order.TotalPrice = price
				}
			case 3:
				// Status: String
				order.Status = m.editBuffer
			}
		}
	case "products":
		if m.cursorRow >= 0 && m.cursorRow < len(m.products) {
			product := &m.products[m.cursorRow]
			switch m.cursorCol {
			case 0:
				// ID: Parse int
				if id, err := parseInt(m.editBuffer); err == nil {
					product.ID = id
				}
			case 1:
				// Name: String
				product.Name = m.editBuffer
			case 2:
				// Price: Parse float
				if price, err := parseFloat(m.editBuffer); err == nil {
					product.Price = price
				}
			case 3:
				// Category: String
				product.Category = m.editBuffer
			}
		}
	case "categories":
		if m.cursorRow >= 0 && m.cursorRow < len(m.categories) {
			category := &m.categories[m.cursorRow]
			switch m.cursorCol {
			case 0:
				// ID: Parse int
				if id, err := parseInt(m.editBuffer); err == nil {
					category.ID = id
				}
			case 1:
				// Name: String
				category.Name = m.editBuffer
			case 2:
				// Slug: String
				category.Slug = m.editBuffer
			}
		}
	}
}

// Helper function to parse int
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// Helper function to parse float
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
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
	// Define headers
	headers := []string{"ID", "USERNAME", "EMAIL", "ACTIVE"}
	
	// Define minimum and ideal column widths
	minColWidths := []int{3, 8, 8, 6} // Minimum acceptable widths
	idealColWidths := []int{5, 20, 30, 10} // Preferred widths when space allows
	
	// Prepare data rows
	rows := make([][]string, len(m.users))
	for i, user := range m.users {
		rows[i] = make([]string, 4)
		rows[i][0] = fmt.Sprintf("%d", user.ID)
		rows[i][1] = user.Username
		rows[i][2] = user.Email
		if user.Active {
			rows[i][3] = "Yes"
		} else {
			rows[i][3] = "No"
		}
	}
	
	return m.renderTable(headers, rows, minColWidths, idealColWidths)
}

func (m model) renderOrdersTable() string {
	// Define headers
	headers := []string{"ID", "USER ID", "TOTAL PRICE", "STATUS"}
	
	// Define minimum and ideal column widths
	minColWidths := []int{3, 6, 8, 8} // Minimum acceptable widths
	idealColWidths := []int{5, 10, 15, 20} // Preferred widths when space allows
	
	// Prepare data rows
	rows := make([][]string, len(m.orders))
	for i, order := range m.orders {
		rows[i] = make([]string, 4)
		rows[i][0] = fmt.Sprintf("%d", order.ID)
		rows[i][1] = fmt.Sprintf("%d", order.UserID)
		rows[i][2] = fmt.Sprintf("$%.2f", order.TotalPrice)
		rows[i][3] = order.Status
	}
	
	return m.renderTable(headers, rows, minColWidths, idealColWidths)
}

func (m model) renderProductsTable() string {
	// Define headers
	headers := []string{"ID", "NAME", "PRICE", "CATEGORY"}
	
	// Define minimum and ideal column widths
	minColWidths := []int{3, 8, 8, 8} // Minimum acceptable widths
	idealColWidths := []int{5, 25, 15, 20} // Preferred widths when space allows
	
	// Prepare data rows
	rows := make([][]string, len(m.products))
	for i, product := range m.products {
		rows[i] = make([]string, 4)
		rows[i][0] = fmt.Sprintf("%d", product.ID)
		rows[i][1] = product.Name
		rows[i][2] = fmt.Sprintf("$%.2f", product.Price)
		rows[i][3] = product.Category
	}
	
	return m.renderTable(headers, rows, minColWidths, idealColWidths)
}

func (m model) renderCategoriesTable() string {
	// Define headers
	headers := []string{"ID", "NAME", "SLUG"}
	
	// Define minimum and ideal column widths
	minColWidths := []int{3, 8, 8} // Minimum acceptable widths
	idealColWidths := []int{5, 25, 25} // Preferred widths when space allows
	
	// Prepare data rows
	rows := make([][]string, len(m.categories))
	for i, category := range m.categories {
		rows[i] = make([]string, 3)
		rows[i][0] = fmt.Sprintf("%d", category.ID)
		rows[i][1] = category.Name
		rows[i][2] = category.Slug
	}
	
	return m.renderTable(headers, rows, minColWidths, idealColWidths)
}

// Helper function to truncate text with ellipsis if needed
func truncateWithEllipsis(text string, width int) string {
	if len(text) <= width {
		return text
	}
	
	if width <= 3 {
		return text[:width]
	}
	
	return text[:width-3] + "..."
}

// Helper function to calculate appropriate column widths based on content and available space
func calculateDynamicWidths(availableWidth int, numColumns int, minWidths []int, idealWidths []int) []int {
	result := make([]int, numColumns)
	
	// Calculate total space needed for borders and padding
	// Each cell has left and right padding (2 chars) and we need space for separators
	tableBorderOverhead := 3 // Left, right border and one extra spacer
	cellPaddingOverhead := 3 * numColumns // Each cell needs padding and separator
	rowNumberColumn := 7 // Width for row numbers column with padding and separator
	
	totalOverhead := tableBorderOverhead + cellPaddingOverhead + rowNumberColumn
	
	usableWidth := availableWidth - totalOverhead
	if usableWidth < 0 {
		usableWidth = 0
	}
	
	// Start with minimum widths
	totalMinWidth := 0
	for i, minWidth := range minWidths {
		result[i] = minWidth
		totalMinWidth += minWidth
	}
	
	// If we have extra space, distribute proportionally based on ideal widths
	if totalMinWidth < usableWidth {
		// Calculate total ideal width
		totalIdealWidth := 0
		for _, idealWidth := range idealWidths {
			totalIdealWidth += idealWidth
		}
		
		extraSpace := usableWidth - totalMinWidth
		spaceUsed := 0
		
		// Distribute extra space proportionally
		for i := 0; i < numColumns; i++ {
			// Calculate proportion of extra space for this column
			proportion := float64(idealWidths[i]) / float64(totalIdealWidth)
			extraForColumn := int(float64(extraSpace) * proportion)
			
			// Don't exceed ideal width
			if result[i]+extraForColumn > idealWidths[i] {
				extraForColumn = idealWidths[i] - result[i]
			}
			
			result[i] += extraForColumn
			spaceUsed += extraForColumn
		}
		
		// If we have any remaining space due to rounding, give it to the last column
		if spaceUsed < extraSpace && numColumns > 0 {
			result[numColumns-1] += (extraSpace - spaceUsed)
		}
	} else if totalMinWidth > usableWidth {
		// We need to reduce widths below minimums
		// Try to keep columns proportional based on ideal widths
		totalIdealWidth := 0
		for _, idealWidth := range idealWidths {
			totalIdealWidth += idealWidth
		}
		
		// Calculate how much we need to reduce by
		reduction := totalMinWidth - usableWidth
		
		// Ensure each column has at least a minimum width (3 chars)
		absoluteMinWidth := 3
		
		// Reduce proportionally
		for i := 0; i < numColumns; i++ {
			proportion := float64(idealWidths[i]) / float64(totalIdealWidth)
			reduceBy := int(float64(reduction) * proportion)
			
			// Ensure we don't go below absolute minimum
			if result[i]-reduceBy < absoluteMinWidth {
				reduceBy = result[i] - absoluteMinWidth
				if reduceBy < 0 {
					reduceBy = 0
				}
			}
			
			result[i] -= reduceBy
		}
	}
	
	return result
}

// Renders a table using lipgloss styles
func (m model) renderTable(headers []string, rows [][]string, minColWidths []int, idealColWidths []int) string {
	// Get available width for the table
	mainBoxWidth := m.width - int(0.2*float64(m.width)) - 4 // Account for sidebar and borders
	if mainBoxWidth < 30 {
		mainBoxWidth = 30 
	}
	
	// Calculate dynamic column widths
	numColumns := len(headers)
	colWidths := calculateDynamicWidths(mainBoxWidth, numColumns, minColWidths, idealColWidths)
	
	var sb strings.Builder
	
	// Style definitions
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333366")).
		Padding(0, 1).
		Align(lipgloss.Left)
	
	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Align(lipgloss.Left)
	
	altRowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#222233")).
		Padding(0, 1).
		Align(lipgloss.Left)
	
	selectedCellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4444AA")).
		Padding(0, 1).
		Align(lipgloss.Left)
	
	editingCellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#AA4444")).
		Padding(0, 1).
		Align(lipgloss.Left)
	
	rowNumStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 1).
		Width(3).
		Align(lipgloss.Right)
	
	// Table border characters for a cleaner look
	borders := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		MiddleLeft:  "├",
		MiddleRight: "┤",
		MiddleTop:   "┬",
		MiddleBottom:"┴",
		Middle:      "┼",
	}
	
	// Render header row
	headerCells := make([]string, len(headers))
	for i, header := range headers {
		// Truncate if necessary
		truncatedHeader := truncateWithEllipsis(header, colWidths[i])
		headerCells[i] = headerStyle.Copy().Width(colWidths[i]).Render(truncatedHeader)
	}
	
	// Add row number header
	rowNumHeader := headerStyle.Copy().Width(3).Render("#")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, rowNumHeader, lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	
	// Render data rows
	dataRows := make([]string, len(rows))
	for i, row := range rows {
		cells := make([]string, len(row))
		
		// Choose style based on row (alternating)
		rowStyle := cellStyle
		if i%2 == 1 {
			rowStyle = altRowStyle
		}
		
		// Format each cell
		for j, cell := range row {
			cellContent := truncateWithEllipsis(cell, colWidths[j])
			
			// Apply appropriate style based on selection/editing state
			styleToUse := rowStyle
			if !m.focusLeft && m.cursorRow == i && m.cursorCol == j {
				if m.editing {
					// Show edit buffer when editing
					editText := truncateWithEllipsis(m.editBuffer, colWidths[j])
					cells[j] = editingCellStyle.Copy().Width(colWidths[j]).Render(editText)
					continue
				} else {
					styleToUse = selectedCellStyle
				}
			}
			cells[j] = styleToUse.Copy().Width(colWidths[j]).Render(cellContent)
		}
		
		// Add row number
		rowNum := rowNumStyle.Copy().Render(fmt.Sprintf("%d", i+1))
		dataRows[i] = lipgloss.JoinHorizontal(lipgloss.Top, rowNum, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	
	// Join all rows with table borders
	tableStyle := lipgloss.NewStyle().
		BorderStyle(borders).
		BorderForeground(lipgloss.Color("#555555"))
	
	tableContent := lipgloss.JoinVertical(lipgloss.Left, append([]string{headerRow}, dataRows...)...)
	table := tableStyle.Render(tableContent)
	
	sb.WriteString(table)
	
	// Show editing help if enabled
	if m.showEditHelp && !m.focusLeft {
		helpStyle := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#999999"))
		
		sb.WriteString("\n")
		sb.WriteString(helpStyle.Render("Navigation: ↑/↓/←/→ or j/k/h/l | Edit: e or Enter | Cancel: Esc"))
		if m.editing {
			sb.WriteString("\n")
			sb.WriteString(helpStyle.Render("Editing: Type to modify | Submit: Enter | Cancel: Esc"))
		}
	}
	
	return sb.String()
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
