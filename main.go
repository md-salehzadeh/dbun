package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/md-salehzadeh/dbun/src/config"
	"github.com/md-salehzadeh/dbun/src/db"
	"github.com/md-salehzadeh/dbun/src/model"
	"github.com/md-salehzadeh/dbun/src/ui"
)

// AppModel represents the application state
type AppModel struct {
	// Database & data
	dbManager      *db.Manager
	dbConfig       config.DBConfig
	tables         []string
	tableData      map[string][]model.RowData
	tableMetadata  map[string][]model.ColumnMetadata
	tableIndices   map[string][]string
	connected      bool
	errorMsg       string

	// UI state
	styles         ui.Styles
	selectedIdx    int
	activeTableIdx int
	mode           model.ViewMode
	width          int
	height         int
	focusLeft      bool
	
	// Scroll state
	sidebarScroll  int
	mainScroll     int

	// Editing state
	editing       bool
	cursorRow     int
	cursorCol     int
	editBuffer    string
	editingField  string
	showEditHelp  bool
	showEditModal bool // New: Flag for modal visibility
	modalTargetRow int // New: Target row for modal edit
	modalTargetCol int // New: Target col for modal edit
}

// Initialize the app model with database connection
func NewAppModel() (AppModel, error) {
	// Load database configuration
	dbConfig := config.LoadConfig()

	m := AppModel{
		dbConfig:      dbConfig,
		tables:        []string{},
		selectedIdx:   0,
		activeTableIdx: 0,
		mode:          model.DataMode,
		width:         80,
		height:        24,
		focusLeft:     true,
		tableData:     make(map[string][]model.RowData),
		tableMetadata: make(map[string][]model.ColumnMetadata),
		tableIndices:  make(map[string][]string),
		connected:     false,
	}

	// Connect to database
	dbm, err := db.NewManager(dbConfig)
	if err != nil {
		return m, err
	}

	m.dbManager = dbm
	m.connected = true

	// Fetch table names
	tables, err := dbm.GetTableNames()
	if err != nil {
		return m, err
	}

	m.tables = tables

	// Pre-fetch metadata for all tables
	for _, table := range tables {
		// Fetch metadata
		metadata, err := dbm.GetTableMetadata(table)
		if err != nil {
			return m, err
		}
		m.tableMetadata[table] = metadata

		// Fetch indices
		indices, err := dbm.GetTableIndices(table)
		if err != nil {
			return m, err
		}
		m.tableIndices[table] = indices

		// Fetch data (limited to 100 rows per table)
		data, err := dbm.GetTableData(table, 100)
		if err != nil {
			return m, err
		}
		m.tableData[table] = data
	}

	return m, nil
}

// Fallback to sample data when database connection fails
func NewAppModelWithSampleData() AppModel {
	// Generate sample data
	users, orders, products, categories := model.GenerateSampleData()

	m := AppModel{
		tables:        []string{"users", "orders", "products", "categories"},
		selectedIdx:   0,
		activeTableIdx: 0,
		mode:          model.DataMode,
		width:         80,
		height:        24,
		focusLeft:     true,
		connected:     true, // Pretend we're connected even though we're using sample data
		tableData:     make(map[string][]model.RowData),
		tableMetadata: make(map[string][]model.ColumnMetadata),
		tableIndices:  make(map[string][]string),
	}

	// Convert sample data to RowData format
	userRows := make([]model.RowData, len(users))
	for i, user := range users {
		userRows[i] = model.RowData{
			"ID":       user.ID,
			"Username": user.Username,
			"Email":    user.Email,
			"Active":   user.Active,
		}
	}
	m.tableData["users"] = userRows
	m.tableMetadata["users"] = []model.ColumnMetadata{
		{Name: "ID", Type: "int", Nullable: false, Key: "PRI"},
		{Name: "Username", Type: "varchar(50)", Nullable: false, Key: "UNI"},
		{Name: "Email", Type: "varchar(100)", Nullable: false, Key: "UNI"},
		{Name: "Active", Type: "tinyint(1)", Nullable: false, Key: ""},
	}
	m.tableIndices["users"] = []string{"PRIMARY", "idx_username", "idx_email"}

	// Convert other sample data to RowData format similarly
	orderRows := make([]model.RowData, len(orders))
	for i, order := range orders {
		orderRows[i] = model.RowData{
			"ID":         order.ID,
			"UserID":     order.UserID,
			"TotalPrice": order.TotalPrice,
			"Status":     order.Status,
		}
	}
	m.tableData["orders"] = orderRows
	m.tableMetadata["orders"] = []model.ColumnMetadata{
		{Name: "ID", Type: "int", Nullable: false, Key: "PRI"},
		{Name: "UserID", Type: "int", Nullable: false, Key: "MUL"},
		{Name: "TotalPrice", Type: "decimal(10,2)", Nullable: false, Key: ""},
		{Name: "Status", Type: "varchar(20)", Nullable: false, Key: ""},
	}
	m.tableIndices["orders"] = []string{"PRIMARY", "idx_user_id"}

	// Add products and categories similarly
	productRows := make([]model.RowData, len(products))
	for i, product := range products {
		productRows[i] = model.RowData{
			"ID":       product.ID,
			"Name":     product.Name,
			"Price":    product.Price,
			"Category": product.Category,
		}
	}
	m.tableData["products"] = productRows
	m.tableMetadata["products"] = []model.ColumnMetadata{
		{Name: "ID", Type: "int", Nullable: false, Key: "PRI"},
		{Name: "Name", Type: "varchar(100)", Nullable: false, Key: ""},
		{Name: "Price", Type: "decimal(10,2)", Nullable: false, Key: ""},
		{Name: "Category", Type: "varchar(50)", Nullable: false, Key: "MUL"},
	}
	m.tableIndices["products"] = []string{"PRIMARY", "idx_category"}

	categoryRows := make([]model.RowData, len(categories))
	for i, category := range categories {
		categoryRows[i] = model.RowData{
			"ID":   category.ID,
			"Name": category.Name,
			"Slug": category.Slug,
		}
	}
	m.tableData["categories"] = categoryRows
	m.tableMetadata["categories"] = []model.ColumnMetadata{
		{Name: "ID", Type: "int", Nullable: false, Key: "PRI"},
		{Name: "Name", Type: "varchar(50)", Nullable: false, Key: ""},
		{Name: "Slug", Type: "varchar(50)", Nullable: false, Key: "UNI"},
	}
	m.tableIndices["categories"] = []string{"PRIMARY", "idx_slug"}

	return m
}

// Clean up resources when the application exits
func (m *AppModel) Close() {
	if m.dbManager != nil {
		m.dbManager.Close()
	}
}

// Tea model interface implementation
func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.styles = ui.NewStyles(m.width, m.height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// Handle key presses based on current state
func (m *AppModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle modal input first if it's active
	if m.showEditModal {
		return m.handleEditModalKeys(msg)
	}

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
		return m.handleEditModeKeys(msg)
	}

	// Handle normal mode keys
	if m.focusLeft {
		return m.handleLeftPanelKeys(msg)
	} else {
		return m.handleRightPanelKeys(msg)
	}
}

// Handle keys when in edit mode
func (m *AppModel) handleEditModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

// Handle keys when focus is on the left panel (table list)
func (m *AppModel) handleLeftPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIdx > 0 {
			m.selectedIdx--
			
			// Adjust scroll position if selection moves out of view
			if m.selectedIdx < m.sidebarScroll {
				m.sidebarScroll = m.selectedIdx
			}
		}
		return m, nil
	case "down", "j":
		if m.selectedIdx < len(m.tables)-1 {
			m.selectedIdx++
			
			// Calculate visible height (approximate based on main height minus borders)
			visibleHeight := m.styles.SidebarStyle.GetHeight() - 2
			
			// Adjust scroll position if selection moves out of view
			if m.selectedIdx >= m.sidebarScroll+visibleHeight {
				m.sidebarScroll = m.selectedIdx - visibleHeight + 1
			}
		}
		return m, nil
	case "enter":
		// Activate the selected table
		m.activeTableIdx = m.selectedIdx
		// Reset main content scroll when changing tables
		m.mainScroll = 0
		m.cursorRow = 0
		m.cursorCol = 0
		return m, nil
	case "pgup":
		// Page up - scroll up by visible height
		visibleHeight := m.styles.SidebarStyle.GetHeight() - 2
		m.sidebarScroll -= visibleHeight
		if m.sidebarScroll < 0 {
			m.sidebarScroll = 0
		}
		
		// Also move selection if it's now out of view
		if m.selectedIdx >= m.sidebarScroll+visibleHeight {
			m.selectedIdx = m.sidebarScroll + visibleHeight - 1
		} else if m.selectedIdx < m.sidebarScroll {
			m.selectedIdx = m.sidebarScroll
		}
		return m, nil
	case "pgdown":
		// Page down - scroll down by visible height
		visibleHeight := m.styles.SidebarStyle.GetHeight() - 2
		maxScroll := len(m.tables) - visibleHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		
		m.sidebarScroll += visibleHeight
		if m.sidebarScroll > maxScroll {
			m.sidebarScroll = maxScroll
		}
		
		// Also move selection if it's now out of view
		if m.selectedIdx < m.sidebarScroll {
			m.selectedIdx = m.sidebarScroll
		} else if m.selectedIdx >= m.sidebarScroll+visibleHeight {
			m.selectedIdx = m.sidebarScroll + visibleHeight - 1
			if m.selectedIdx >= len(m.tables) {
				m.selectedIdx = len(m.tables) - 1
			}
		}
		return m, nil
	case "home":
		// Scroll to top
		m.sidebarScroll = 0
		if m.selectedIdx < m.sidebarScroll {
			m.selectedIdx = m.sidebarScroll
		}
		return m, nil
	case "end":
		// Scroll to bottom
		visibleHeight := m.styles.SidebarStyle.GetHeight() - 2
		maxScroll := len(m.tables) - visibleHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		m.sidebarScroll = maxScroll
		
		if m.selectedIdx < m.sidebarScroll {
			m.selectedIdx = m.sidebarScroll
		}
		return m, nil
	}
	
	return m, nil
}

// Handle keys when focus is on the right panel (table data)
func (m *AppModel) handleRightPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Tab switching
	switch msg.String() {
	case "d":
		m.mode = model.DataMode
		m.mainScroll = 0 // Reset scroll position when changing view
		return m, nil
	case "s":
		m.mode = model.StructureMode
		m.mainScroll = 0 // Reset scroll position when changing view
		return m, nil
	case "i":
		m.mode = model.IndicesMode
		m.mainScroll = 0 // Reset scroll position when changing view
		return m, nil
	}

	// Common scrolling keys for all modes
	switch msg.String() {
	case "pgup":
		// Page up - scroll up by visible height
		visibleHeight := m.styles.MainBoxStyle.GetHeight() - 2
		m.mainScroll -= visibleHeight
		if m.mainScroll < 0 {
			m.mainScroll = 0
		}
		return m, nil
	case "pgdown":
		var maxContent int
		if m.activeTableIdx >= 0 && m.activeTableIdx < len(m.tables) {
			table := m.tables[m.activeTableIdx]
			// Determine content height based on mode
			if m.mode == model.DataMode && m.tableData[table] != nil {
				maxContent = len(m.tableData[table]) + 1 // +1 for header row
			} else if m.mode == model.StructureMode && m.tableMetadata[table] != nil {
				maxContent = len(m.tableMetadata[table]) + 2 // +2 for title and blank line
			} else if m.mode == model.IndicesMode && m.tableIndices[table] != nil {
				maxContent = len(m.tableIndices[table]) + 2 // +2 for title and blank line
			}
		}
		
		visibleHeight := m.styles.MainBoxStyle.GetHeight() - 2
		m.mainScroll += visibleHeight
		
		// Calculate max scroll position
		maxScroll := maxContent - visibleHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		
		if m.mainScroll > maxScroll {
			m.mainScroll = maxScroll
		}
		return m, nil
	case "home":
		m.mainScroll = 0
		return m, nil
	case "end":
		var maxContent int
		if m.activeTableIdx >= 0 && m.activeTableIdx < len(m.tables) {
			table := m.tables[m.activeTableIdx]
			// Determine content height based on mode
			if m.mode == model.DataMode && m.tableData[table] != nil {
				maxContent = len(m.tableData[table]) + 1 // +1 for header row
			} else if m.mode == model.StructureMode && m.tableMetadata[table] != nil {
				maxContent = len(m.tableMetadata[table]) + 2 // +2 for title and blank line
			} else if m.mode == model.IndicesMode && m.tableIndices[table] != nil {
				maxContent = len(m.tableIndices[table]) + 2 // +2 for title and blank line
			}
		}
		
		visibleHeight := m.styles.MainBoxStyle.GetHeight() - 2
		maxScroll := maxContent - visibleHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		
		m.mainScroll = maxScroll
		return m, nil
	}

	// Data mode specific navigation and actions
	if m.mode == model.DataMode {
		switch msg.String() {
		case "up", "k":
			if m.cursorRow > 0 {
				m.cursorRow--
				
				// Adjust scroll if cursor moves out of view
				if m.cursorRow < m.mainScroll {
					m.mainScroll = m.cursorRow
				}
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
				
				// Calculate visible height (approximate based on main height minus borders and header)
				visibleHeight := m.styles.MainBoxStyle.GetHeight() - 3
				
				// Adjust scroll if cursor moves out of view
				if m.cursorRow >= m.mainScroll+visibleHeight {
					m.mainScroll = m.cursorRow - visibleHeight + 1
				}
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
			// Enter INLINE edit mode for the current cell
			return m.enterInlineEditMode(), nil
		case "ctrl+e":
			// Enter MODAL edit mode for the current cell
			return m.enterModalEditMode(), nil
		case "ctrl+n":
			return m.setCellToNull(), nil
		}
	}
	
	return m, nil
}

// Enter INLINE edit mode for the current cell
func (m *AppModel) enterInlineEditMode() tea.Model {
	// Only allow editing in data mode
	if m.mode != model.DataMode {
		return m
	}
	
	// Set editing flag and prepare edit buffer
	m.editing = true
	m.showEditModal = false // Ensure modal is not shown
	m.editBuffer = m.getCurrentCellValue()
	m.editingField = m.getCurrentFieldName()
	
	return m
}


// Enter MODAL edit mode for the current cell
func (m *AppModel) enterModalEditMode() tea.Model {
	// Only allow editing in data mode
	if m.mode != model.DataMode {
		return m
	}

	// Set editing and modal flags, prepare buffer, store target
	m.editing = true
	m.showEditModal = true
	m.editBuffer = m.getCurrentCellValue()
	m.editingField = m.getCurrentFieldName()
	m.modalTargetRow = m.cursorRow // Store original cursor position
	m.modalTargetCol = m.cursorCol

	return m
}

// Handle keys when in modal edit mode
func (m *AppModel) handleEditModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel modal edit
		m.editing = false
		m.showEditModal = false
		m.editBuffer = ""
		return m, nil

	case "enter":
		// Submit modal edit
		m.applyEdit() // applyEdit will now use modalTargetRow/Col
		m.editing = false
		m.showEditModal = false
		m.editBuffer = ""
		return m, nil

	case "backspace":
		// Delete last character in modal
		if len(m.editBuffer) > 0 {
			// Handle multi-byte characters correctly if necessary (simple approach here)
			// For simplicity, assuming simple runes or ASCII
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}
		return m, nil

	default:
		// Add character to modal edit buffer
		// Consider adding rune handling for broader character support
		if len(msg.String()) == 1 { // Basic check, might need refinement for complex inputs
			m.editBuffer += msg.String()
		}
		return m, nil
	}
}

// Get the name of the current field being edited
func (m *AppModel) getCurrentFieldName() string {
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

// Get the current value of the selected cell
func (m *AppModel) getCurrentCellValue() string {
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
					// Return empty string for NULL values instead of "NULL"
					return "" 
				}
				
				// Use FormatValue from model package for consistency
				return model.FormatValue(val)
			}
		}
	}
	
	return ""
}

// Apply the edit to the current cell
func (m *AppModel) applyEdit() {
	targetRow := m.cursorRow
	targetCol := m.cursorCol

	// If the edit came from the modal, use the stored target coordinates
	if m.showEditModal {
		targetRow = m.modalTargetRow
		targetCol = m.modalTargetCol
	}


	if m.activeTableIdx < 0 || m.activeTableIdx >= len(m.tables) {
		return
	}
	
	table := m.tables[m.activeTableIdx]
	
	// Get data and metadata
	data, dataOk := m.tableData[table]
	metadata, metaOk := m.tableMetadata[table]
	
	// Use targetRow/targetCol for validation and access
	if !dataOk || !metaOk || targetRow >= len(data) || targetCol >= len(metadata) {
		return
	}
	
	colMeta := metadata[targetCol] // Get the specific column metadata
	colName := colMeta.Name
	colType := colMeta.Type
	isNullable := colMeta.Nullable
	
	// Get the row
	row := data[targetRow]
	
	// Parse the edited value based on column type
	var newValue interface{}
	
	// Handle empty input for nullable columns -> treat as empty string
	if m.editBuffer == "" && isNullable {
		newValue = "" 
	} else {
		// Parse based on data type (existing logic)
		if strings.Contains(colType, "int") {
			if val, err := model.ParseInt(m.editBuffer); err == nil {
				newValue = val
			} else if isNullable { 
				// If parsing fails and column is nullable, set to nil (or keep original?)
				// Let's keep nil for now as empty string doesn't make sense for int
				newValue = nil 
			} else {
				newValue = row[colName] // Keep original if not nullable and parse fails
			}
		} else if strings.Contains(colType, "float") || 
				strings.Contains(colType, "double") || 
				strings.Contains(colType, "decimal") {
			if val, err := model.ParseFloat(m.editBuffer); err == nil {
				newValue = val
			} else if isNullable {
				newValue = nil // Keep nil for parse errors on nullable numerics
			} else {
				newValue = row[colName]
			}
		} else if strings.Contains(colType, "bool") || 
				strings.Contains(colType, "tinyint(1)") {
			lowerBuffer := strings.ToLower(m.editBuffer)
			if lowerBuffer == "true" || lowerBuffer == "yes" || lowerBuffer == "1" {
				newValue = true
			} else if lowerBuffer == "false" || lowerBuffer == "no" || lowerBuffer == "0" {
				newValue = false
			} else if isNullable { 
				newValue = nil // Keep nil for parse errors on nullable bools
			} else { 
				newValue = false // Default to false for non-nullable bools on parse error
			}
		} else {
			// Default to string for other types (VARCHAR, TEXT, etc.)
			// Empty string for nullable is handled above.
			// If the buffer is empty and not nullable, set empty string
			if m.editBuffer == "" && !isNullable {
				newValue = "" // Set non-nullable strings to empty string if buffer is empty
			} else {
				// Treat the buffer content as a string, including "NULL" or "null"
				newValue = m.editBuffer
			}
		}
	}
	
	// Update the value in memory (not in the database)
	row[colName] = newValue
	data[targetRow] = row // Use targetRow
	m.tableData[table] = data
}

// New function to handle setting cell to NULL
func (m *AppModel) setCellToNull() tea.Model {
	// Only works in data mode and when right panel has focus
	if m.mode != model.DataMode || m.focusLeft {
		return m
	}

	if m.activeTableIdx < 0 || m.activeTableIdx >= len(m.tables) {
		return m
	}
	
	table := m.tables[m.activeTableIdx]
	
	// Get data and metadata
	data, dataOk := m.tableData[table]
	metadata, metaOk := m.tableMetadata[table]
	
	// Validate cursor position
	if !dataOk || !metaOk || m.cursorRow >= len(data) || m.cursorCol >= len(metadata) {
		return m
	}
	
	colMeta := metadata[m.cursorCol]
	
	// Check if the column is nullable
	if colMeta.Nullable {
		colName := colMeta.Name
		row := data[m.cursorRow]
		
		// Set the value to nil directly in memory
		row[colName] = nil
		data[m.cursorRow] = row
		m.tableData[table] = data
		
		// Optionally: Add feedback to the user (e.g., status message)
		// m.statusMessage = fmt.Sprintf("Cell [%d, %d] set to NULL", m.cursorRow, m.cursorCol)
	} else {
		// Optionally: Add feedback if column is not nullable
		// m.statusMessage = fmt.Sprintf("Column '%s' is not nullable", colMeta.Name)
	}

	return m
}

func (m AppModel) View() string {
	// Check if connected to database
	if (!m.connected) {
		return fmt.Sprintf("Not connected to database: %s\nPress Ctrl+C to exit", m.errorMsg)
	}

	// Ensure we have valid dimensions
	if (m.width == 0 || m.height == 0) {
		return "Loading..."
	}

	// Initialize styles if they haven't been initialized yet
	// We'll check if the width of our styles is 0, which would indicate they're not initialized
	styles := m.styles
	if styles.MainBoxStyle.GetWidth() <= 0 {
		styles = ui.NewStyles(m.width, m.height)
	}

	// Update styles based on current focus and mode
	styles = styles.UpdateStyles(m.focusLeft, m.mode)

	// Calculate layout elements size
	sidebarWidth := int(0.2 * float64(m.width))
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}
	mainBoxWidth := m.width - sidebarWidth - 4 // Account for borders

	// Render table list sidebar
	sidebarView := ui.RenderTableList(styles, m.tables, m.selectedIdx, m.activeTableIdx, m.sidebarScroll)

	// Render main content based on the selected table and mode
	var mainContent string
	
	if m.activeTableIdx < 0 || m.activeTableIdx >= len(m.tables) {
		mainContent = "No table selected"
	} else {
		activeTable := m.tables[m.activeTableIdx]
		
		if m.mode == model.DataMode {
			// Display table data
			mainContent = ui.RenderTableData(
				styles, 
				mainBoxWidth,
				activeTable,
				m.tableMetadata[activeTable],
				m.tableData[activeTable],
				m.cursorRow,
				m.cursorCol,
				m.focusLeft,
				m.editing,
				m.editBuffer,
				m.mainScroll,
			)
		} else if m.mode == model.StructureMode {
			// Display table structure
			mainContent = ui.RenderTableStructure(activeTable, m.tableMetadata[activeTable], m.mainScroll)
		} else if m.mode == model.IndicesMode {
			// Display indices information
			mainContent = ui.RenderTableIndices(activeTable, m.tableIndices[activeTable], m.mainScroll)
		}
	}
	
	mainBoxView := styles.MainBoxStyle.Render(mainContent)

	// Render the tab bar
	buttonBar := lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 0, 0, 2).
		Background(lipgloss.Color("#222222"))

	buttons := lipgloss.JoinHorizontal(lipgloss.Top,
		styles.DataTabStyle.Render("Data"),
		styles.StructureTabStyle.Render("Structure"),
		styles.IndicesTabStyle.Render("Indices"),
	)
	
	buttonSection := buttonBar.Render(buttons)
	
	// Combine the main views
	layout := lipgloss.JoinVertical(lipgloss.Top, 
		buttonSection,
		lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainBoxView),
	)

	var doc strings.Builder
	doc.WriteString(layout)

	// Add status bar
	statusBar := ui.RenderStatusBar(styles, m.width)
	doc.WriteString("\n" + statusBar)

	// Add help text if enabled
	if m.showEditHelp && !m.focusLeft {
		helpStyle := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#999999"))
		
		doc.WriteString("\n")
		// Updated help text to use Ctrl+N
		doc.WriteString(helpStyle.Render("Navigation: ↑/↓/←/→ or j/k/h/l | Edit: e/Enter | Modal Edit: Ctrl+E | Set Null: Ctrl+N")) 
		doc.WriteString("\n")
		doc.WriteString(helpStyle.Render("Scroll: PgUp/PgDn/Home/End | Switch View: d/s/i | Toggle Help: ? | Quit: q/Ctrl+C")) // Added more help
		if m.editing {
			doc.WriteString("\n")
			doc.WriteString(helpStyle.Render("Editing: Type to modify | Submit: Enter | Cancel: Esc"))
		}
	}

	// Render the modal if active
	if m.showEditModal {
		modalView := ui.RenderEditModal(styles, m.width, m.height, m.editingField, m.editBuffer)
		// Place the modal centered on top of the existing layout
		// We need to join the layout and modal correctly. Lipgloss Place is good for this.
		finalView := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalView, lipgloss.WithWhitespaceChars(" "))
		return finalView // Return the view with the modal overlay
	}


	return doc.String()
}

func main() {
	// Load configuration
	dbConfig := config.LoadConfig()
	
	fmt.Printf("Connecting to MySQL database at %s:%d with user %s and database %s\n", 
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Database)
	
	// Initialize the model with database connection
	m, err := NewAppModel()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		fmt.Println("Falling back to sample data...")
		
		// Fallback to sample data if database connection fails
		m = NewAppModelWithSampleData()
	}
	
	// Make sure to clean up on exit
	defer m.Close()
	
	// Start the application
	p := tea.NewProgram(&m, tea.WithAltScreen())
	
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
