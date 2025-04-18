package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Database connection settings
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// Table metadata
type TableMetadata struct {
	Name    string
	Columns []ColumnMetadata
}

// Column metadata
type ColumnMetadata struct {
	Name     string
	Type     string
	Nullable bool
	Key      string
}

// Generic row data (for any table)
type RowData map[string]interface{}

// App model
type model struct {
	db             *sql.DB       // Database connection
	dbConfig       DBConfig      // Database configuration
	tables         []string      // List of tables
	selectedIdx    int           // Currently selected table index
	activeTableIdx int           // Currently active table index
	mode           string        // Mode: "Data", "Structure", "Indices"
	width          int           // Terminal width
	height         int           // Terminal height
	focusLeft      bool          // Focus on left (true) or right (false) box
	
	// Dynamic table data
	tableData      map[string][]RowData       // Data for each table
	tableMetadata  map[string][]ColumnMetadata // Metadata for each table
	tableIndices   map[string][]string         // Indices for each table
	
	// Editing state
	editing       bool   // Whether we're in editing mode
	cursorRow     int    // Current cursor row in the table
	cursorCol     int    // Current cursor column in the table
	editBuffer    string // Buffer for the current edit
	editingField  string // Name of the field being edited
	showEditHelp  bool   // Whether to show editing help
	
	// Connection state
	connected     bool   // Whether connected to database
	errorMsg      string // Error message if any
}

// Initialize the database connection
func initDB(config DBConfig) (*sql.DB, error) {
	// Format DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", 
		config.User, config.Password, config.Host, config.Port, config.Database)
	
	fmt.Printf("DSN: %s\n", dsn)
	
	// Open connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	
	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}
	
	return db, nil
}

// Fetches all table names from the database
func fetchTableNames(db *sql.DB) ([]string, error) {
	query := "SHOW TABLES"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching tables: %v", err)
	}
	defer rows.Close()
	
	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %v", err)
		}
		tables = append(tables, tableName)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %v", err)
	}
	
	return tables, nil
}

// Fetches column metadata for a specific table
func fetchTableMetadata(db *sql.DB, tableName string) ([]ColumnMetadata, error) {
	query := fmt.Sprintf("DESCRIBE %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching table metadata: %v", err)
	}
	defer rows.Close()
	
	var columns []ColumnMetadata
	for rows.Next() {
		var field, dataType, null, key, defaultVal, extra sql.NullString
		if err := rows.Scan(&field, &dataType, &null, &key, &defaultVal, &extra); err != nil {
			return nil, fmt.Errorf("error scanning column metadata: %v", err)
		}
		
		column := ColumnMetadata{
			Name: field.String,
			Type: dataType.String,
			Nullable: null.String == "YES",
			Key: key.String,
		}
		columns = append(columns, column)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %v", err)
	}
	
	return columns, nil
}

// Fetches indices for a specific table
func fetchTableIndices(db *sql.DB, tableName string) ([]string, error) {
	query := fmt.Sprintf("SHOW INDEX FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching table indices: %v", err)
	}
	defer rows.Close()
	
	// Get column names from result set
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting index columns: %v", err)
	}
	
	// Create a slice of interface{} to hold the values
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	
	indexMap := make(map[string]bool) // Use map to avoid duplicates
	
	// For each row, scan all columns
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, fmt.Errorf("error scanning index data: %v", err)
		}
		
		// Extract the key name (usually in position 2)
		keyNameIdx := 2 // Default position for key_name
		
		// Find the key_name column position for flexibility
		for i, colName := range columns {
			if strings.EqualFold(colName, "Key_name") {
				keyNameIdx = i
				break
			}
		}
		
		// Get the key name if it exists and is not null
		if keyNameIdx < len(values) && values[keyNameIdx] != nil {
			switch v := values[keyNameIdx].(type) {
			case []byte:
				indexMap[string(v)] = true
			case string:
				indexMap[v] = true
			}
		}
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating indices: %v", err)
	}
	
	// Convert map keys to slice
	indices := make([]string, 0, len(indexMap))
	for key := range indexMap {
		indices = append(indices, key)
	}
	
	return indices, nil
}

// Fetches data for a specific table (limited to a reasonable number of rows)
func fetchTableData(db *sql.DB, tableName string, limit int) ([]RowData, error) {
	// Get columns first to handle the results properly
	columns, err := fetchTableMetadata(db, tableName)
	if err != nil {
		return nil, err
	}
	
	// Build query with column names with proper backtick escaping for MySQL
	columnNames := make([]string, len(columns))
	for i, col := range columns {
		// Escape column names with backticks to handle reserved words and special characters
		columnNames[i] = fmt.Sprintf("`%s`", col.Name)
	}
	
	query := fmt.Sprintf("SELECT %s FROM `%s` LIMIT %d", 
		strings.Join(columnNames, ", "), tableName, limit)
	
	fmt.Printf("Running query: %s\n", query)
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer rows.Close()
	
	// Get column types to properly handle NULL values
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("error getting column types: %v", err)
	}
	
	var result []RowData
	
	// For each row
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columnNames))
		// Create a slice of pointers to the values
		scanArgs := make([]interface{}, len(columnNames))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		
		// Scan the row into the slice of interface{}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		
		// Create a map to hold the row data
		rowData := make(RowData)
		
		// For each column in the row
		for i, col := range columns {
			colName := col.Name
			
			// Handle different types
			var v interface{}
			val := values[i]
			
			// Handle NULL values
			if val == nil {
				v = nil
			} else {
				// Handle different types based on the column type
				switch colTypes[i].DatabaseTypeName() {
				case "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT":
					switch val.(type) {
					case []byte:
						v, _ = strconv.Atoi(string(val.([]byte)))
					default:
						v = val
					}
				case "DECIMAL", "FLOAT", "DOUBLE":
					switch val.(type) {
					case []byte:
						v, _ = strconv.ParseFloat(string(val.([]byte)), 64)
					default:
						v = val
					}
				default:
					// For TEXT, VARCHAR, etc. convert []byte to string
					switch val.(type) {
					case []byte:
						v = string(val.([]byte))
					default:
						v = val
					}
				}
			}
			
			rowData[colName] = v
		}
		
		result = append(result, rowData)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}
	
	return result, nil
}

// Initialize the model with database connection and data
func initModel(dbConfig DBConfig) (model, error) {
	m := model{
		dbConfig:      dbConfig,
		tables:        []string{},
		selectedIdx:   0,
		activeTableIdx: 0,
		mode:          "Data",
		width:         80,
		height:        24,
		focusLeft:     true,
		tableData:     make(map[string][]RowData),
		tableMetadata: make(map[string][]ColumnMetadata),
		tableIndices:  make(map[string][]string),
		connected:     false,
	}
	
	// Connect to database
	db, err := initDB(dbConfig)
	if err != nil {
		return m, err
	}
	
	m.db = db
	m.connected = true
	
	// Fetch table names
	tables, err := fetchTableNames(db)
	if err != nil {
		return m, err
	}
	
	m.tables = tables
	
	// Pre-fetch metadata for all tables
	for _, table := range tables {
		// Fetch metadata
		metadata, err := fetchTableMetadata(db, table)
		if err != nil {
			return m, err
		}
		m.tableMetadata[table] = metadata
		
		// Fetch indices
		indices, err := fetchTableIndices(db, table)
		if err != nil {
			return m, err
		}
		m.tableIndices[table] = indices
		
		// Fetch data (limited to 100 rows per table)
		data, err := fetchTableData(db, table, 100)
		if err != nil {
			return m, err
		}
		m.tableData[table] = data
	}
	
	return m, nil
}

// Sample data structures (kept for reference/fallback)
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

// Initialize sample data for tables - only used as fallback if database connection fails
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
					if m.activeTableIdx >= 0 && m.activeTableIdx < len(m.tables) {
						table := m.tables[m.activeTableIdx]
						if data, ok := m.tableData[table]; ok {
							maxRows = len(data)
						}
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
					if m.activeTableIdx >= 0 && m.activeTableIdx < len(m.tables) {
						table := m.tables[m.activeTableIdx]
						if metadata, ok := m.tableMetadata[table]; ok {
							maxCols = len(metadata)
						}
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
	if m.activeTableIdx < 0 || m.activeTableIdx >= len(m.tables) {
		return ""
	}
	
	table := m.tables[m.activeTableIdx]
	
	// Get column names from metadata
	if metadata, ok := m.tableMetadata[table]; ok && m.cursorCol < len(metadata) {
		return metadata[m.cursorCol].Name
	}
	
	return ""
}

// Gets the current value of the selected cell
func (m *model) getCurrentCellValue() string {
	if m.activeTableIdx < 0 || m.activeTableIdx >= len(m.tables) {
		return ""
	}
	
	table := m.tables[m.activeTableIdx]
	
	// Get value from tableData
	if data, ok := m.tableData[table]; ok && m.cursorRow < len(data) {
		row := data[m.cursorRow]
		
		// Get column name from metadata
		if metadata, ok := m.tableMetadata[table]; ok && m.cursorCol < len(metadata) {
			colName := metadata[m.cursorCol].Name
			
			// Format the value based on its type
			if val, ok := row[colName]; ok {
				if val == nil {
					return "NULL"
				}
				
				switch v := val.(type) {
				case bool:
					if v {
						return "Yes"
					}
					return "No"
				case int, int8, int16, int32, int64:
					return fmt.Sprintf("%d", v)
				case float32, float64:
					return fmt.Sprintf("%.2f", v)
				default:
					return fmt.Sprintf("%v", v)
				}
			}
		}
	}
	
	return ""
}

// Apply the edit to the current cell
func (m *model) applyEdit() {
	if m.activeTableIdx < 0 || m.activeTableIdx >= len(m.tables) {
		return
	}
	
	table := m.tables[m.activeTableIdx]
	
	// Get data and metadata
	data, dataOk := m.tableData[table]
	metadata, metaOk := m.tableMetadata[table]
	
	if !dataOk || !metaOk || m.cursorRow >= len(data) || m.cursorCol >= len(metadata) {
		return
	}
	
	colName := metadata[m.cursorCol].Name
	colType := metadata[m.cursorCol].Type
	
	// Get the row
	row := data[m.cursorRow]
	
	// Parse the edited value based on column type
	var newValue interface{}
	
	// Handle NULL value
	if m.editBuffer == "NULL" || m.editBuffer == "null" {
		newValue = nil
	} else {
		// Parse based on data type
		if strings.Contains(colType, "int") {
			if val, err := parseInt(m.editBuffer); err == nil {
				newValue = val
			}
		} else if strings.Contains(colType, "float") || 
				strings.Contains(colType, "double") || 
				strings.Contains(colType, "decimal") {
			if val, err := parseFloat(m.editBuffer); err == nil {
				newValue = val
			}
		} else if strings.Contains(colType, "bool") || 
				strings.Contains(colType, "tinyint(1)") {
			newValue = m.editBuffer == "Yes" || m.editBuffer == "yes" || 
					  m.editBuffer == "true" || m.editBuffer == "TRUE" || 
					  m.editBuffer == "1"
		} else {
			// Default to string for other types
			newValue = m.editBuffer
		}
	}
	
	// Update the value in memory (not in the database)
	row[colName] = newValue
	data[m.cursorRow] = row
	m.tableData[table] = data
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
	// Check if connected to database
	if (!m.connected) {
		return fmt.Sprintf("Not connected to database: %s\nPress Ctrl+C to exit", m.errorMsg)
	}

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
	
	if m.activeTableIdx < 0 || m.activeTableIdx >= len(m.tables) {
		mainContent = "No table selected"
	} else {
		activeTable := m.tables[m.activeTableIdx]
		
		if m.mode == "Data" {
			// Display table data from database
			mainContent = m.renderTableData(activeTable)
		} else if m.mode == "Structure" {
			// Display table structure from metadata
			mainContent = m.renderTableStructure(activeTable)
		} else if m.mode == "Indices" {
			// Display indices information from metadata
			mainContent = m.renderTableIndices(activeTable)
		} else {
			mainContent = fmt.Sprintf("Unknown mode: %s", m.mode)
		}
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

// Renders table data from the database
func (m model) renderTableData(tableName string) string {
	// Ensure we have metadata and data
	metadata, metaOk := m.tableMetadata[tableName]
	data, dataOk := m.tableData[tableName]
	
	if !metaOk || !dataOk {
		return fmt.Sprintf("No data available for table: %s", tableName)
	}
	
	// Define headers from metadata
	headers := make([]string, len(metadata))
	for i, col := range metadata {
		headers[i] = strings.ToUpper(col.Name)
	}
	
	// Define minimum and ideal column widths
	// Use a reasonable default for all columns
	minColWidths := make([]int, len(headers))
	idealColWidths := make([]int, len(headers))
	
	for i, col := range metadata {
		// Set minimum width based on header length
		headerLen := len(col.Name)
		if headerLen < 3 {
			minColWidths[i] = 3
		} else {
			minColWidths[i] = headerLen
		}
		
		// Set ideal width based on data type
		if strings.Contains(col.Type, "int") {
			idealColWidths[i] = 8
		} else if strings.Contains(col.Type, "float") || 
				strings.Contains(col.Type, "double") || 
				strings.Contains(col.Type, "decimal") {
			idealColWidths[i] = 12
		} else if strings.Contains(col.Type, "varchar") {
			// Extract size from varchar(N)
			size := 20 // Default
			if start := strings.Index(col.Type, "("); start != -1 {
				if end := strings.Index(col.Type[start:], ")"); end != -1 {
					if num, err := strconv.Atoi(col.Type[start+1 : start+end]); err == nil {
						size = num
						if size > 30 {
							size = 30 // Cap at 30 for display
						}
					}
				}
			}
			idealColWidths[i] = size
		} else if strings.Contains(col.Type, "text") {
			idealColWidths[i] = 30
		} else {
			idealColWidths[i] = 15
		}
	}
	
	// Prepare data rows
	rows := make([][]string, len(data))
	for i, row := range data {
		rows[i] = make([]string, len(headers))
		
		for j, col := range metadata {
			colName := col.Name
			
			// Format the value based on its type
			if val, ok := row[colName]; ok {
				if val == nil {
					rows[i][j] = "NULL"
				} else {
					switch v := val.(type) {
					case bool:
						if v {
							rows[i][j] = "Yes"
						} else {
							rows[i][j] = "No"
						}
					case int, int8, int16, int32, int64:
						rows[i][j] = fmt.Sprintf("%d", v)
					case float32, float64:
						rows[i][j] = fmt.Sprintf("%.2f", v)
					default:
						rows[i][j] = fmt.Sprintf("%v", v)
					}
				}
			} else {
				rows[i][j] = ""
			}
		}
	}
	
	return m.renderTable(headers, rows, minColWidths, idealColWidths)
}

// Renders table structure from metadata
func (m model) renderTableStructure(tableName string) string {
	metadata, ok := m.tableMetadata[tableName]
	if (!ok) {
		return fmt.Sprintf("No metadata available for table: %s", tableName)
	}
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s Structure:\n\n", tableName))
	
	for _, col := range metadata {
		nullableStr := "NOT NULL"
		if col.Nullable {
			nullableStr = "NULL"
		}
		
		keyStr := ""
		if col.Key != "" {
			keyStr = fmt.Sprintf(" (%s)", col.Key)
		}
		
		sb.WriteString(fmt.Sprintf("%s: %s %s%s\n", col.Name, col.Type, nullableStr, keyStr))
	}
	
	return sb.String()
}

// Renders table indices information
func (m model) renderTableIndices(tableName string) string {
	indices, ok := m.tableIndices[tableName]
	if (!ok) {
		return fmt.Sprintf("No index information available for table: %s", tableName)
	}
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s Indices:\n\n", tableName))
	
	for _, idx := range indices {
		sb.WriteString(fmt.Sprintf("Index: %s\n", idx))
	}
	
	return sb.String()
}

// Helper methods to render table data
func (m model) renderUsersTable() string {
	// Define headers
	headers := []string{"ID", "USERNAME", "EMAIL", "ACTIVE"}
	
	// Define minimum and ideal column widths
	minColWidths := []int{3, 8, 8, 6} // Minimum acceptable widths
	idealColWidths := []int{5, 20, 30, 10} // Preferred widths when space allows
	
	// Prepare data rows
	rows := make([][]string, len(m.tableData["users"]))
	for i, user := range m.tableData["users"] {
		rows[i] = make([]string, 4)
		rows[i][0] = fmt.Sprintf("%d", user["ID"])
		rows[i][1] = user["Username"].(string)
		rows[i][2] = user["Email"].(string)
		if user["Active"].(bool) {
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
	rows := make([][]string, len(m.tableData["orders"]))
	for i, order := range m.tableData["orders"] {
		rows[i] = make([]string, 4)
		rows[i][0] = fmt.Sprintf("%d", order["ID"])
		rows[i][1] = fmt.Sprintf("%d", order["UserID"])
		rows[i][2] = fmt.Sprintf("$%.2f", order["TotalPrice"])
		rows[i][3] = order["Status"].(string)
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
	rows := make([][]string, len(m.tableData["products"]))
	for i, product := range m.tableData["products"] {
		rows[i] = make([]string, 4)
		rows[i][0] = fmt.Sprintf("%d", product["ID"])
		rows[i][1] = product["Name"].(string)
		rows[i][2] = fmt.Sprintf("$%.2f", product["Price"])
		rows[i][3] = product["Category"].(string)
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
	rows := make([][]string, len(m.tableData["categories"]))
	for i, category := range m.tableData["categories"] {
		rows[i] = make([]string, 3)
		rows[i][0] = fmt.Sprintf("%d", category["ID"])
		rows[i][1] = category["Name"].(string)
		rows[i][2] = category["Slug"].(string)
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
	// Default database config
	config := DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "", // You should use environment variables in real applications
		Database: "test",
	}
	
	// Check environment variables for configuration
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Host = host
	}
	
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}
	
	if user := os.Getenv("DB_USER"); user != "" {
		config.User = user
	}
	
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Password = password
	}
	
	if database := os.Getenv("DB_NAME"); database != "" {
		config.Database = database
	}
	
	fmt.Printf("Connecting to MySQL database at %s:%d with user %s and database %s\n", 
		config.Host, config.Port, config.User, config.Database)
	
	// Initialize the model with database connection
	m, err := initModel(config)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		fmt.Println("Falling back to sample data...")
		
		// Fallback to sample data if database connection fails
		users, orders, products, categories := initSampleData()
		m = model{
			tables:         []string{"users", "orders", "products", "categories"},
			selectedIdx:    0,
			activeTableIdx: 0,
			mode:           "Data",
			width:          80,
			height:         24,
			focusLeft:      true,
			connected:      true, // Pretend we're connected even though we're using sample data
			tableData:      make(map[string][]RowData),
			tableMetadata:  make(map[string][]ColumnMetadata),
			tableIndices:   make(map[string][]string),
		}
		
		// Convert sample data to RowData format
		userRows := make([]RowData, len(users))
		for i, user := range users {
			userRows[i] = RowData{
				"ID":       user.ID,
				"Username": user.Username,
				"Email":    user.Email,
				"Active":   user.Active,
			}
		}
		m.tableData["users"] = userRows
		m.tableMetadata["users"] = []ColumnMetadata{
			{Name: "ID", Type: "int", Nullable: false, Key: "PRI"},
			{Name: "Username", Type: "varchar(50)", Nullable: false, Key: "UNI"},
			{Name: "Email", Type: "varchar(100)", Nullable: false, Key: "UNI"},
			{Name: "Active", Type: "tinyint(1)", Nullable: false, Key: ""},
		}
		m.tableIndices["users"] = []string{"PRIMARY", "idx_username", "idx_email"}
		
		// Convert other sample data to RowData format similarly
		orderRows := make([]RowData, len(orders))
		for i, order := range orders {
			orderRows[i] = RowData{
				"ID":         order.ID,
				"UserID":     order.UserID,
				"TotalPrice": order.TotalPrice,
				"Status":     order.Status,
			}
		}
		m.tableData["orders"] = orderRows
		m.tableMetadata["orders"] = []ColumnMetadata{
			{Name: "ID", Type: "int", Nullable: false, Key: "PRI"},
			{Name: "UserID", Type: "int", Nullable: false, Key: "MUL"},
			{Name: "TotalPrice", Type: "decimal(10,2)", Nullable: false, Key: ""},
			{Name: "Status", Type: "varchar(20)", Nullable: false, Key: ""},
		}
		m.tableIndices["orders"] = []string{"PRIMARY", "idx_user_id"}
		
		// Add products and categories similarly
		productRows := make([]RowData, len(products))
		for i, product := range products {
			productRows[i] = RowData{
				"ID":       product.ID,
				"Name":     product.Name,
				"Price":    product.Price,
				"Category": product.Category,
			}
		}
		m.tableData["products"] = productRows
		m.tableMetadata["products"] = []ColumnMetadata{
			{Name: "ID", Type: "int", Nullable: false, Key: "PRI"},
			{Name: "Name", Type: "varchar(100)", Nullable: false, Key: ""},
			{Name: "Price", Type: "decimal(10,2)", Nullable: false, Key: ""},
			{Name: "Category", Type: "varchar(50)", Nullable: false, Key: "MUL"},
		}
		m.tableIndices["products"] = []string{"PRIMARY", "idx_category"}
		
		categoryRows := make([]RowData, len(categories))
		for i, category := range categories {
			categoryRows[i] = RowData{
				"ID":   category.ID,
				"Name": category.Name,
				"Slug": category.Slug,
			}
		}
		m.tableData["categories"] = categoryRows
		m.tableMetadata["categories"] = []ColumnMetadata{
			{Name: "ID", Type: "int", Nullable: false, Key: "PRI"},
			{Name: "Name", Type: "varchar(50)", Nullable: false, Key: ""},
			{Name: "Slug", Type: "varchar(50)", Nullable: false, Key: "UNI"},
		}
		m.tableIndices["categories"] = []string{"PRIMARY", "idx_slug"}
	}
	
	// Start the application
	p := tea.NewProgram(&m, tea.WithAltScreen())
	
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
