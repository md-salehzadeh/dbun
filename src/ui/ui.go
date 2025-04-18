package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/md-salehzadeh/dbun/src/model"
)

// Styles holds all the styling for the application
type Styles struct {
	// Border colors
	ActiveBorderColor   lipgloss.Color
	InactiveBorderColor lipgloss.Color

	// Sidebar styles
	SidebarStyle      lipgloss.Style
	SelectedItemStyle lipgloss.Style
	ActiveItemStyle   lipgloss.Style
	NormalItemStyle   lipgloss.Style

	// Main content styles
	MainBoxStyle lipgloss.Style

	// Tab styles
	TabStyle            lipgloss.Style
	DataTabStyle        lipgloss.Style
	StructureTabStyle   lipgloss.Style
	IndicesTabStyle     lipgloss.Style

	// Table styles
	HeaderStyle        lipgloss.Style
	CellStyle          lipgloss.Style
	AltRowStyle        lipgloss.Style
	SelectedCellStyle  lipgloss.Style
	EditingCellStyle   lipgloss.Style
	RowNumStyle        lipgloss.Style
	TableBorders       lipgloss.Border

	// Status bar styles
	StatusBarStyle  lipgloss.Style
	StatusStyle     lipgloss.Style
	EncodingStyle   lipgloss.Style
	FishCakeStyle   lipgloss.Style
	StatusTextStyle lipgloss.Style
}

// NewStyles creates the default UI styles
func NewStyles(width, height int) Styles {
	// Calculate dynamic widths and heights
	sidebarWidth := int(0.2 * float64(width))
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}

	buttonHeightWithPadding := 3
	statusBarHeight := 1
	mainHeight := height - buttonHeightWithPadding - statusBarHeight - 2

	mainBoxWidth := width - sidebarWidth - 4 // Account for borders
	if mainBoxWidth < 20 {
		mainBoxWidth = 20
	}

	buttonWidth := int(float64(width-6) / 3.0)

	// Define colors
	activeBorderColor := lipgloss.Color("#FF00FF")   // Magenta
	inactiveBorderColor := lipgloss.Color("#444444") // Grey

	// Table borders
	borders := lipgloss.Border{
		Top:          "─",
		Bottom:       "─",
		Left:         "│",
		Right:        "│",
		TopLeft:      "┌",
		TopRight:     "┐",
		BottomLeft:   "└",
		BottomRight:  "┘",
		MiddleLeft:   "├",
		MiddleRight:  "┤",
		MiddleTop:    "┬",
		MiddleBottom: "┴",
		Middle:       "┼",
	}

	return Styles{
		ActiveBorderColor:   activeBorderColor,
		InactiveBorderColor: inactiveBorderColor,

		// Sidebar
		SidebarStyle: lipgloss.NewStyle().
			Width(sidebarWidth).
			Height(mainHeight).
			Border(lipgloss.RoundedBorder()).
			Align(lipgloss.Left),

		SelectedItemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF00FF")),
		ActiveItemStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#444444")).
			Foreground(lipgloss.Color("#FFFFFF")),
		NormalItemStyle: lipgloss.NewStyle(),

		// Main Box
		MainBoxStyle: lipgloss.NewStyle().
			Width(mainBoxWidth).
			Height(mainHeight).
			Border(lipgloss.RoundedBorder()).
			Align(lipgloss.Left),

		// Tabs
		TabStyle: lipgloss.NewStyle().
			Padding(1, 2).
			Bold(true).
			Width(buttonWidth),

		// Table styles
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333366")).
			Padding(0, 1).
			Align(lipgloss.Left),

		CellStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Align(lipgloss.Left),

		AltRowStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#222233")).
			Padding(0, 1).
			Align(lipgloss.Left),

		SelectedCellStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#4444AA")).
			Padding(0, 1).
			Align(lipgloss.Left),

		EditingCellStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#AA4444")).
			Padding(0, 1).
			Align(lipgloss.Left),

		RowNumStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1).
			Width(3).
			Align(lipgloss.Right),

		TableBorders: borders,

		// Status bar
		StatusBarStyle: lipgloss.NewStyle().
			Width(width).
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"}),

		StatusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1),

		EncodingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#A550DF")).
			Padding(0, 1),

		FishCakeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6124DF")).
			Padding(0, 1),

		StatusTextStyle: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"}),
	}
}

// UpdateStyles updates the styles based on focus and mode
func (s Styles) UpdateStyles(focusLeft bool, mode model.ViewMode) Styles {
	newStyles := s

	// Update border styles based on focus
	if focusLeft {
		newStyles.SidebarStyle = newStyles.SidebarStyle.BorderForeground(newStyles.ActiveBorderColor)
		newStyles.MainBoxStyle = newStyles.MainBoxStyle.BorderForeground(newStyles.InactiveBorderColor)
	} else {
		newStyles.SidebarStyle = newStyles.SidebarStyle.BorderForeground(newStyles.InactiveBorderColor)
		newStyles.MainBoxStyle = newStyles.MainBoxStyle.BorderForeground(newStyles.ActiveBorderColor)
	}

	// Update tab styles based on current mode
	newStyles.DataTabStyle = newStyles.TabStyle.Copy().Foreground(newStyles.InactiveBorderColor)
	newStyles.StructureTabStyle = newStyles.TabStyle.Copy().Foreground(newStyles.InactiveBorderColor)
	newStyles.IndicesTabStyle = newStyles.TabStyle.Copy().Foreground(newStyles.InactiveBorderColor)

	switch mode {
	case model.DataMode:
		newStyles.DataTabStyle = newStyles.TabStyle.Copy().
			Foreground(newStyles.ActiveBorderColor).
			Underline(true)
	case model.StructureMode:
		newStyles.StructureTabStyle = newStyles.TabStyle.Copy().
			Foreground(newStyles.ActiveBorderColor).
			Underline(true)
	case model.IndicesMode:
		newStyles.IndicesTabStyle = newStyles.TabStyle.Copy().
			Foreground(newStyles.ActiveBorderColor).
			Underline(true)
	}

	return newStyles
}

// RenderTableList renders the list of tables with selection indicators
func RenderTableList(styles Styles, tables []string, selectedIdx, activeTableIdx int) string {
	var sidebar strings.Builder

	for i, table := range tables {
		var line string
		cursor := " " // Default cursor

		if selectedIdx == i && activeTableIdx == i {
			cursor = ">"
			line = styles.ActiveItemStyle.Render(fmt.Sprintf("%s %s", cursor, table))
		} else if selectedIdx == i {
			cursor = ">"
			line = styles.SelectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, table))
		} else if activeTableIdx == i {
			line = styles.ActiveItemStyle.Render(fmt.Sprintf("%s %s", cursor, table))
		} else {
			line = styles.NormalItemStyle.Render(fmt.Sprintf("%s %s", cursor, table))
		}

		sidebar.WriteString(line + "\n")
	}

	return styles.SidebarStyle.Render(sidebar.String())
}

// RenderStatusBar renders the application status bar
func RenderStatusBar(styles Styles, width int) string {
	w := lipgloss.Width

	statusKey := styles.StatusStyle.Render("STATUS")
	encoding := styles.EncodingStyle.Render("UTF-8")
	fishCake := styles.FishCakeStyle.Render("🍥 Fish Cake")
	statusVal := styles.StatusTextStyle.
		Width(width - w(statusKey) - w(encoding) - w(fishCake) - 5).
		Render("Ravishing")

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
		encoding,
		fishCake,
	)

	return styles.StatusBarStyle.Render(bar)
}

// CalculateDynamicWidths calculates appropriate column widths based on content and available space
func CalculateDynamicWidths(availableWidth, numColumns int, minWidths, idealWidths []int) []int {
	result := make([]int, numColumns)
	
	// Calculate total space needed for borders and padding
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

// RenderTable renders a data table with headers and rows
func RenderTable(styles Styles, mainBoxWidth int, 
                headers []string, rows [][]string, 
                minColWidths, idealColWidths []int,
                cursorRow, cursorCol int, 
                focusLeft, editing bool,
                editBuffer string) string {
	
	var sb strings.Builder
	
	// Calculate dynamic column widths
	numColumns := len(headers)
	colWidths := CalculateDynamicWidths(mainBoxWidth, numColumns, minColWidths, idealColWidths)
	
	// Render header row
	headerCells := make([]string, len(headers))
	for i, header := range headers {
		truncatedHeader := model.TruncateWithEllipsis(header, colWidths[i])
		headerCells[i] = styles.HeaderStyle.Copy().Width(colWidths[i]).Render(truncatedHeader)
	}
	
	// Add row number header
	rowNumHeader := styles.HeaderStyle.Copy().Width(3).Render("#")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, rowNumHeader, lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	
	// Render data rows
	dataRows := make([]string, len(rows))
	for i, row := range rows {
		cells := make([]string, len(row))
		
		// Choose style based on row (alternating)
		rowStyle := styles.CellStyle
		if i%2 == 1 {
			rowStyle = styles.AltRowStyle
		}
		
		// Format each cell
		for j, cell := range row {
			cellContent := model.TruncateWithEllipsis(cell, colWidths[j])
			
			// Apply appropriate style based on selection/editing state
			styleToUse := rowStyle
			if !focusLeft && cursorRow == i && cursorCol == j {
				if editing {
					// Show edit buffer when editing
					editText := model.TruncateWithEllipsis(editBuffer, colWidths[j])
					cells[j] = styles.EditingCellStyle.Copy().Width(colWidths[j]).Render(editText)
					continue
				} else {
					styleToUse = styles.SelectedCellStyle
				}
			}
			cells[j] = styleToUse.Copy().Width(colWidths[j]).Render(cellContent)
		}
		
		// Add row number
		rowNum := styles.RowNumStyle.Copy().Render(fmt.Sprintf("%d", i+1))
		dataRows[i] = lipgloss.JoinHorizontal(lipgloss.Top, rowNum, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	
	// Join all rows with table borders
	tableStyle := lipgloss.NewStyle().
		BorderStyle(styles.TableBorders).
		BorderForeground(lipgloss.Color("#555555"))
	
	tableContent := lipgloss.JoinVertical(lipgloss.Left, append([]string{headerRow}, dataRows...)...)
	table := tableStyle.Render(tableContent)
	
	sb.WriteString(table)
	
	return sb.String()
}

// RenderTableData formats table data into a displayable format
func RenderTableData(styles Styles, mainBoxWidth int, 
                     tableName string, 
                     metadata []model.ColumnMetadata, 
                     data []model.RowData,
                     cursorRow, cursorCol int, 
                     focusLeft, editing bool,
                     editBuffer string) string {
	
	if len(metadata) == 0 || len(data) == 0 {
		return fmt.Sprintf("No data available for table: %s", tableName)
	}
	
	// Create headers from metadata
	headers := make([]string, len(metadata))
	for i, col := range metadata {
		headers[i] = strings.ToUpper(col.Name)
	}
	
	// Define minimum and ideal column widths
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
					if num, err := model.ParseInt(col.Type[start+1 : start+end]); err == nil {
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
	
	return RenderTable(styles, mainBoxWidth, headers, rows, minColWidths, idealColWidths, 
	                   cursorRow, cursorCol, focusLeft, editing, editBuffer)
}

// RenderTableStructure renders a table's structure information
func RenderTableStructure(tableName string, metadata []model.ColumnMetadata) string {
	if len(metadata) == 0 {
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

// RenderTableIndices renders a table's indices information
func RenderTableIndices(tableName string, indices []string) string {
	if len(indices) == 0 {
		return fmt.Sprintf("No index information available for table: %s", tableName)
	}
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s Indices:\n\n", tableName))
	
	for _, idx := range indices {
		sb.WriteString(fmt.Sprintf("Index: %s\n", idx))
	}
	
	return sb.String()
}