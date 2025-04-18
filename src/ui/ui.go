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
			BorderForeground(inactiveBorderColor).
			Padding(0, 1).
			MarginRight(1).
			Align(lipgloss.Left),

		SelectedItemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF00FF")), // Magenta
		ActiveItemStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#444444")).
			Foreground(lipgloss.Color("#FFFFFF")),
		NormalItemStyle: lipgloss.NewStyle(),

		// Main Box
		MainBoxStyle: lipgloss.NewStyle().
			Width(mainBoxWidth).
			Height(mainHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(inactiveBorderColor).
			Padding(0, 1).
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
			Width(4).
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
func RenderTableList(styles Styles, tables []string, selectedIdx, activeTableIdx int, scrollPosition int) string {
	// Calculate inner height available for list items
	// Overhead: TopBorder(1), BottomBorder(1), Title(1), Blank after Title(1), TopScrollIndicator(1/0), BottomScrollIndicator(1/0), Pagination(1/0), Blank before Pagination(1/0)
	boxInnerHeight := styles.SidebarStyle.GetHeight() - 2 // Account for top/bottom border
	if boxInnerHeight < 0 { boxInnerHeight = 0 }

	titleHeight := 2 // Title line + blank line after
	scrollIndicatorHeight := 1 // Each indicator takes 1 line

	// Calculate maxVisibleItems based on available space
	availableHeight := boxInnerHeight - titleHeight
	if scrollPosition > 0 {
		availableHeight -= scrollIndicatorHeight // Space for "↑ Previous"
	}

	// Estimate footer height to reserve space
	estimatedFooterHeight := 0
	// Temporarily calculate max items without footer to see if footer is needed
	tempMaxItems := availableHeight
	if tempMaxItems < 0 { tempMaxItems = 0 }
	tempEndIdx := scrollPosition + tempMaxItems
	if tempEndIdx < len(tables) {
		estimatedFooterHeight += scrollIndicatorHeight // "↓ More"
	}
	if len(tables) > tempMaxItems { // If total items > estimated visible
		if estimatedFooterHeight > 0 { // Add blank line before pagination if "↓ More" is shown
			estimatedFooterHeight += 1
		}
		estimatedFooterHeight += 1 // Pagination line
	}
	availableHeight -= estimatedFooterHeight

	maxVisibleItems := availableHeight
	if maxVisibleItems < 0 { maxVisibleItems = 0 } // Cannot be negative


	// Calculate which portion of the list to show
	endIdx := scrollPosition + maxVisibleItems
	if endIdx > len(tables) {
		endIdx = len(tables)
	}
	// Ensure start index is valid
	if scrollPosition < 0 { scrollPosition = 0 }
	if scrollPosition > len(tables) { scrollPosition = len(tables) } // Can be empty if scrolled past end
	if endIdx < scrollPosition { endIdx = scrollPosition } // Ensure end is not before start


	// Prepare content with proper spacing and alignment
	var content strings.Builder
	var currentContentHeight int

	// Add title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#6A0DAD")). // Purple color for title
		Width(styles.SidebarStyle.GetWidth() - 4). // Account for border padding
		Align(lipgloss.Center)

	content.WriteString(titleStyle.Render("TABLES"))
	content.WriteString("\n") // Blank line after title
	currentContentHeight += 2

	// Add scroll indicator if needed
	if scrollPosition > 0 {
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Align(lipgloss.Center).
			Width(styles.SidebarStyle.GetWidth() - 4)

		content.WriteString(indicatorStyle.Render("↑ Previous"))
		content.WriteString("\n")
		currentContentHeight += 1
	}

	// Show visible table entries
	numItemsRendered := 0
	for i := scrollPosition; i < endIdx; i++ {
		var lineStyle lipgloss.Style
		cursor := " " // Default cursor

		// Determine style based on selection and active state
		itemStyleWidth := styles.SidebarStyle.GetWidth() - 6 // Account for border, padding, and cursor
		if itemStyleWidth < 1 { itemStyleWidth = 1 }

		if selectedIdx == i && activeTableIdx == i {
			cursor = "●" // Active and selected
			lineStyle = styles.ActiveItemStyle.Copy().Bold(true).Width(itemStyleWidth)
		} else if selectedIdx == i {
			cursor = ">" // Just selected
			lineStyle = styles.SelectedItemStyle.Copy().Bold(true).Width(itemStyleWidth)
		} else if activeTableIdx == i {
			cursor = " " // Just active (can happen if selection moved away?) - Use ActiveItemStyle
			lineStyle = styles.ActiveItemStyle.Copy().Width(itemStyleWidth)
		} else {
			cursor = " " // Normal item
			lineStyle = styles.NormalItemStyle.Copy().Width(itemStyleWidth)
		}

		// Create a fixed-width table name
		tableName := tables[i]
		// Truncate based on the style's width calculation
		// Subtract cursor width (1) and space (1)
		maxTextWidth := itemStyleWidth - 2
		if maxTextWidth < 0 { maxTextWidth = 0 }
		truncatedTableName := model.TruncateWithEllipsis(tableName, maxTextWidth)

		// Render the line
		content.WriteString(fmt.Sprintf("%s %s\n", cursor, lineStyle.Render(truncatedTableName)))
		currentContentHeight += 1
		numItemsRendered++
	}

	// --- Footer Section ---
	var footerBuilder strings.Builder
	footerHeight := 0

	// Add scroll indicator if there are more items below
	showMoreIndicator := endIdx < len(tables)
	if showMoreIndicator {
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Align(lipgloss.Center).
			Width(styles.SidebarStyle.GetWidth() - 4)

		footerBuilder.WriteString(indicatorStyle.Render("↓ More") + "\n")
		footerHeight += 1
	}

	// Add pagination info
	// Show pagination if total items > number actually rendered OR if we are scrolled
	showPagination := len(tables) > numItemsRendered || scrollPosition > 0
	if showPagination && len(tables) > 0 { // Also check if there are any tables
		paginationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Align(lipgloss.Center).
			Width(styles.SidebarStyle.GetWidth() - 4)

		paginationText := fmt.Sprintf("%d-%d of %d",
			scrollPosition+1,
			min(scrollPosition+numItemsRendered, len(tables)), // Use actual rendered count
			len(tables))

		// Add blank line before pagination only if "↓ More" indicator is also shown
		if showMoreIndicator {
			footerBuilder.WriteString("\n")
			footerHeight += 1
		}
		footerBuilder.WriteString(paginationStyle.Render(paginationText)) // No newline needed if it's the last line
		footerHeight += 1
	}
	footerContent := footerBuilder.String()

	// --- Combine and Pad ---
	// Calculate remaining space to fill
	remainingHeight := boxInnerHeight - currentContentHeight - footerHeight
	if remainingHeight < 0 { remainingHeight = 0 }

	// Add blank lines between items and footer
	content.WriteString(strings.Repeat("\n", remainingHeight))

	// Add the footer content
	content.WriteString(footerContent)

	// Render the final sidebar content within its style
	// Note: SidebarStyle already includes Padding(0, 1)
	finalContentStr := content.String()
	// Ensure the final string doesn't exceed the inner height due to edge cases
	finalLines := strings.Split(finalContentStr, "\n")
	if len(finalLines) > boxInnerHeight {
		finalContentStr = strings.Join(finalLines[:boxInnerHeight], "\n")
	}


	return styles.SidebarStyle.Render(finalContentStr)
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
	rowNumHeader := styles.HeaderStyle.Copy().Width(4).Render("#")
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
		
		// Add row number with consistent width
		rowNum := styles.RowNumStyle.Copy().Width(4).Render(fmt.Sprintf("%d", i+1))
		dataRows[i] = lipgloss.JoinHorizontal(lipgloss.Top, rowNum, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	
	// Join all rows with table borders
	tableStyle := lipgloss.NewStyle().
		BorderStyle(styles.TableBorders).
		BorderForeground(lipgloss.Color("#555555"))
	
	// Add the table content
	if len(dataRows) > 0 {
		tableContent := lipgloss.JoinVertical(lipgloss.Left, append([]string{headerRow}, dataRows...)...)
		table := tableStyle.Render(tableContent)
		sb.WriteString(table)
	} else {
		// Handle empty data set with just headers
		tableContent := headerRow
		table := tableStyle.Render(tableContent)
		sb.WriteString(table)
		sb.WriteString("\nNo data to display")
	}
	
	return sb.String()
}

// RenderTableData formats table data into a displayable format with scrolling
func RenderTableData(styles Styles, mainBoxWidth int,
                     tableName string,
                     metadata []model.ColumnMetadata,
                     data []model.RowData,
                     cursorRow, cursorCol int,
                     focusLeft, editing bool,
                     editBuffer string,
                     scrollPosition int) string {

	// Calculate inner height available for rows
	boxInnerHeight := styles.MainBoxStyle.GetHeight() - 2 // Subtract top/bottom border of MainBoxStyle
	if boxInnerHeight < 0 { boxInnerHeight = 0 }
	// Overhead: Title(1), Blank(2), TableHeader(1), TableBorders(2), Blank(1), ScrollInfo(1) = 8 lines
	fixedOverhead := 8
	maxVisibleRows := boxInnerHeight - fixedOverhead
	if maxVisibleRows < 0 { // Ensure it's not negative
		maxVisibleRows = 0
	}

	// --- Title ---
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#1E90FF")).
		Padding(0, 1).
		Align(lipgloss.Center).
		Width(mainBoxWidth - 4) // Account for MainBoxStyle padding

	titleContent := titleStyle.Render(tableName) + "\n\n" // Title + 2 blank lines
	currentContentHeight := 3

	// --- Handle No Metadata or No Data ---
	if len(metadata) == 0 {
		message := "No structure available for this table."
		noDataStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Align(lipgloss.Center).Width(mainBoxWidth - 4)
		messageContent := noDataStyle.Render(message)
		currentContentHeight += 1 // For the message line

		blanksNeeded := boxInnerHeight - currentContentHeight
		if blanksNeeded < 0 { blanksNeeded = 0 }
		return titleContent + messageContent + strings.Repeat("\n", blanksNeeded)
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
		headerLen := len(col.Name)
		minColWidths[i] = max(3, headerLen) // Min width is 3 or header length

		// Ideal width logic (simplified)
		idealColWidths[i] = 15 // Default
		if strings.Contains(col.Type, "int") { idealColWidths[i] = 8 }
		if strings.Contains(col.Type, "float") || strings.Contains(col.Type, "double") || strings.Contains(col.Type, "decimal") { idealColWidths[i] = 12 }
		if strings.Contains(col.Type, "varchar") { idealColWidths[i] = 25 } // Adjust as needed
		if strings.Contains(col.Type, "text") { idealColWidths[i] = 30 }
		idealColWidths[i] = max(minColWidths[i], idealColWidths[i]) // Ideal >= Min
	}


	if len(data) == 0 {
		// Render empty table (just header) + "No data" message
		emptyRows := [][]string{}
		tableContent := RenderTable(styles, mainBoxWidth - 4, headers, emptyRows,
			minColWidths, idealColWidths,
			-1, -1, focusLeft, false, "",
		)
		noDataStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Align(lipgloss.Center).Width(mainBoxWidth - 4)
		noDataMessage := noDataStyle.Render("No data to display")

		// Calculate height: Table(Header(1)+Borders(2)) + Blank(1) + NoDataMsg(1) = 5 lines
		tableSectionHeight := 5
		currentContentHeight += tableSectionHeight

		blanksNeeded := boxInnerHeight - currentContentHeight
		if blanksNeeded < 0 { blanksNeeded = 0 }

		return titleContent + tableContent + "\n" + noDataMessage + strings.Repeat("\n", blanksNeeded)
	}


	// --- Apply Scrolling ---
	visibleData := data
	if scrollPosition < 0 { scrollPosition = 0 } // Ensure scroll is not negative

	endPos := scrollPosition + maxVisibleRows
	if endPos > len(data) {
		endPos = len(data)
	}

	// Adjust scrollPosition if it's past the end
	if scrollPosition >= len(data) {
		scrollPosition = max(0, len(data)-maxVisibleRows) // Go to last possible page
		endPos = len(data)
	}

	// Ensure end is after start
	if endPos < scrollPosition { endPos = scrollPosition }

	visibleData = data[scrollPosition:endPos]
	numVisibleRows := len(visibleData)


	// --- Prepare Data Rows ---
	rows := make([][]string, numVisibleRows)
	for i, rowData := range visibleData {
		rows[i] = make([]string, len(headers))
		for j, col := range metadata {
			colName := col.Name
			if val, ok := rowData[colName]; ok {
				rows[i][j] = model.FormatValue(val) // Use a helper for consistent formatting
			} else {
				rows[i][j] = ""
			}
		}
	}

	// --- Adjust Cursor ---
	adjustedCursorRow := cursorRow - scrollPosition
	if adjustedCursorRow < 0 || adjustedCursorRow >= numVisibleRows {
		adjustedCursorRow = -1 // Cursor is not visible
	}


	// --- Render Table ---
	tableContent := RenderTable(styles, mainBoxWidth - 4, headers, rows,
		minColWidths, idealColWidths,
		adjustedCursorRow, cursorCol,
		focusLeft, editing, editBuffer)
	// Table height: Header(1) + Rows(numVisibleRows) + Borders(2) = numVisibleRows + 3
	tableHeight := numVisibleRows + 3
	currentContentHeight += tableHeight


	// --- Footer / Scroll Info ---
	var footerBuilder strings.Builder
	footerHeight := 0

	scrollInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Align(lipgloss.Left)

	totalRows := len(data)
	currentStart := scrollPosition + 1
	currentEnd := scrollPosition + numVisibleRows

	// Show scroll info if needed
	showScrollInfo := totalRows > maxVisibleRows || scrollPosition > 0
	if showScrollInfo {
		var indicators []string
		if scrollPosition > 0 { indicators = append(indicators, "↑ Prev") }
		if currentEnd < totalRows { indicators = append(indicators, "↓ More") }

		paginationInfo := fmt.Sprintf("Rows %d-%d of %d", currentStart, currentEnd, totalRows)

		scrollInfo := paginationInfo
		if len(indicators) > 0 {
			scrollInfo = strings.Join(indicators, " | ") + "   " + paginationInfo
		}
		footerBuilder.WriteString(scrollInfoStyle.Render(scrollInfo))
		footerHeight = 1 // Scroll info takes 1 line
	} else {
		// If no scroll info, we still need a line for consistent height calculation (blank line after table)
		footerHeight = 1
	}
	currentContentHeight += footerHeight // Account for the line after table (either scroll info or blank)


	// --- Combine and Pad ---
	var finalContent strings.Builder
	finalContent.WriteString(titleContent)
	finalContent.WriteString(tableContent)
	finalContent.WriteString("\n") // Blank line OR line where footer starts
	finalContent.WriteString(footerBuilder.String())

	// Add padding lines if needed
	blanksNeeded := boxInnerHeight - currentContentHeight
	if blanksNeeded > 0 {
		finalContent.WriteString(strings.Repeat("\n", blanksNeeded))
	}

	// Final safety check: Truncate if somehow too long
	finalStr := finalContent.String()
	finalLines := strings.Split(finalStr, "\n")
	if len(finalLines) > boxInnerHeight {
		finalStr = strings.Join(finalLines[:boxInnerHeight], "\n")
	}

	return finalStr
}

// RenderTableStructure renders a table's structure information with scrolling
func RenderTableStructure(tableName string, metadata []model.ColumnMetadata, scrollPosition int) string {
	if len(metadata) == 0 {
		return fmt.Sprintf("No metadata available for table: %s", tableName)
	}
	
	// Prepare content
	var sb strings.Builder
	
	// Add title with consistent styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#1E90FF")).
		Padding(0, 1).
		Align(lipgloss.Center)
		
	sb.WriteString(titleStyle.Render(tableName))
	sb.WriteString("\n\n")
	
	// Style for column names
	colNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#AACCFF"))
		
	// Style for types
	typeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
		
	// Style for constraints
	constraintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFCCAA"))
		
	// Style for keys
	keyStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("#AAFFAA"))
	
	// Calculate visible rows based on reasonable estimate
	visibleMetadata := metadata
	maxVisibleRows := 20 // Approximate - could refine based on box height
	
	if scrollPosition >= 0 && len(metadata) > maxVisibleRows {
		endPos := scrollPosition + maxVisibleRows
		if endPos > len(metadata) {
			endPos = len(metadata)
		}
		
		if scrollPosition < len(metadata) {
			visibleMetadata = metadata[scrollPosition:endPos]
		} else {
			visibleMetadata = []model.ColumnMetadata{}
		}
	}
	
	// Add scroll indicators with consistent styling
	if scrollPosition > 0 {
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Align(lipgloss.Center)
			
		sb.WriteString(indicatorStyle.Render("↑ Previous columns"))
		sb.WriteString("\n\n")
	}
	
	// Structure table header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333366"))
		
	sb.WriteString(headerStyle.Render(" Column Name       Type                    Constraints     Key "))
	sb.WriteString("\n")
	sb.WriteString(headerStyle.Render("─────────────────────────────────────────────────────────────"))
	sb.WriteString("\n")
	
	// Show the column details
	for i, col := range visibleMetadata {
		// Background color for alternating rows
		rowStyle := lipgloss.NewStyle()
		if i%2 == 1 {
			rowStyle = rowStyle.Background(lipgloss.Color("#222233"))
		}
		
		nullableStr := "NOT NULL"
		if col.Nullable {
			nullableStr = "NULL"
		}
		
		keyStr := col.Key
		if keyStr == "" {
			keyStr = "-"
		}
		
		// Format each field with padding to align columns
		colNameText := rowStyle.Render(" " + colNameStyle.Render(model.TruncateWithEllipsis(col.Name, 16)))
		typeText := rowStyle.Render(" " + typeStyle.Render(model.TruncateWithEllipsis(col.Type, 22)))
		nullText := rowStyle.Render(" " + constraintStyle.Render(model.TruncateWithEllipsis(nullableStr, 14)))
		keyText := rowStyle.Render(" " + keyStyle.Render(model.TruncateWithEllipsis(keyStr, 3)) + " ")
		
		sb.WriteString(colNameText + typeText + nullText + keyText)
		sb.WriteString("\n")
	}
	
	// Add more indicator and pagination info
	if scrollPosition + len(visibleMetadata) < len(metadata) {
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Align(lipgloss.Center)
			
		sb.WriteString("\n")
		sb.WriteString(indicatorStyle.Render("↓ More columns"))
	}
	
	// Add scroll position indicator with nice styling
	if len(metadata) > maxVisibleRows {
		paginationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Align(lipgloss.Right).
			PaddingTop(1)
			
		paginationText := fmt.Sprintf("Columns %d-%d of %d", 
			scrollPosition+1, 
			scrollPosition+len(visibleMetadata), 
			len(metadata))
			
		sb.WriteString("\n")
		sb.WriteString(paginationStyle.Render(paginationText))
	}
	
	return sb.String()
}

// RenderTableIndices renders a table's indices information with scrolling
func RenderTableIndices(tableName string, indices []string, scrollPosition int) string {
	if len(indices) == 0 {
		return fmt.Sprintf("No index information available for table: %s", tableName)
	}
	
	// Prepare content
	var sb strings.Builder
	
	// Add title with consistent styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#1E90FF")).
		Padding(0, 1).
		Align(lipgloss.Center)
		
	sb.WriteString(titleStyle.Render(tableName))
	sb.WriteString("\n\n")
	
	// Calculate visible indices
	visibleIndices := indices
	maxVisibleRows := 20 // Approximate - could refine based on box height
	
	if scrollPosition >= 0 && len(indices) > maxVisibleRows {
		endPos := scrollPosition + maxVisibleRows
		if endPos > len(indices) {
			endPos = len(indices)
		}
		
		if scrollPosition < len(indices) {
			visibleIndices = indices[scrollPosition:endPos]
		} else {
			visibleIndices = []string{}
		}
	}
	
	// Add scroll indicators with consistent styling
	if scrollPosition > 0 {
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Align(lipgloss.Center)
			
		sb.WriteString(indicatorStyle.Render("↑ Previous indices"))
		sb.WriteString("\n\n")
	}
	
	// Table header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333366"))
		
	sb.WriteString(headerStyle.Render(" Index Name                       Type        Columns                   "))
	sb.WriteString("\n")
	sb.WriteString(headerStyle.Render("─────────────────────────────────────────────────────────────────────"))
	sb.WriteString("\n")
	
	// Index type and name styling
	indexNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#AACCFF"))
		
	indexTypeStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("#AAFFAA"))
	
	// For demo, we'll parse the index string to extract type and involved columns
	// In a real scenario, you might have more structured index data
	for i, idx := range visibleIndices {
		// Background color for alternating rows
		rowStyle := lipgloss.NewStyle()
		if i%2 == 1 {
			rowStyle = rowStyle.Background(lipgloss.Color("#222233"))
		}
		
		// Extract index type (simple demo parsing)
		indexName := idx
		indexType := "INDEX"
		
		if strings.HasPrefix(strings.ToUpper(idx), "PRIMARY") {
			indexType = "PRIMARY"
		} else if strings.HasPrefix(strings.ToUpper(idx), "UNIQUE") {
			indexType = "UNIQUE"
		} else if strings.HasPrefix(strings.ToUpper(idx), "IDX_") || strings.HasPrefix(idx, "index_") {
			indexType = "INDEX"
		}
		
		// Extract columns (simplified - in real app you would have actual index metadata with columns)
		columns := "N/A"
		if strings.Contains(idx, "_") {
			parts := strings.Split(idx, "_")
			if len(parts) > 1 {
				columns = strings.Join(parts[1:], ", ")
			}
		}
		
		// Format fields with proper alignment and styling
		nameText := rowStyle.Render(" " + indexNameStyle.Render(model.TruncateWithEllipsis(indexName, 30)))
		typeText := rowStyle.Render(" " + indexTypeStyle.Render(model.TruncateWithEllipsis(indexType, 10)))
		columnsText := rowStyle.Render(" " + model.TruncateWithEllipsis(columns, 25) + " ")
		
		sb.WriteString(nameText + typeText + columnsText)
		sb.WriteString("\n")
	}
	
	// Add more indicator and pagination info
	if scrollPosition + len(visibleIndices) < len(indices) {
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Align(lipgloss.Center)
			
		sb.WriteString("\n")
		sb.WriteString(indicatorStyle.Render("↓ More indices"))
	}
	
	// Add scroll position indicator with nice styling
	if len(indices) > maxVisibleRows {
		paginationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Align(lipgloss.Right).
			PaddingTop(1)
			
		paginationText := fmt.Sprintf("Indices %d-%d of %d", 
			scrollPosition+1, 
			scrollPosition+len(visibleIndices), 
			len(indices))
			
		sb.WriteString("\n")
		sb.WriteString(paginationStyle.Render(paginationText))
	}
	
	return sb.String()
}

// Helper function to find minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to find maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RenderEditModal renders a floating modal for editing cell data
func RenderEditModal(styles Styles, termWidth, termHeight int, fieldName, editBuffer string) string {
	// Define modal dimensions (relative to terminal size)
	modalWidth := min(termWidth-10, 60) // Max 60 chars wide, or less if terminal is small
	// Simple height for now, could be dynamic later
	modalHeight := 7

	// Style for the modal box
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.DoubleBorder(), true).
		BorderForeground(styles.ActiveBorderColor). // Use active color
		Padding(1, 2).
		Background(lipgloss.Color("#333333")) // Dark background

	// Title
	title := fmt.Sprintf("Edit %s", fieldName)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	titleLine := titleStyle.Render(model.TruncateWithEllipsis(title, modalWidth-4)) // Truncate title if needed

	// Edit area - show the buffer content
	// Simple single-line editor for now
	editAreaStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#444444")). // Slightly different background for input
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	// Display buffer with a cursor indicator (simple pipe char)
	// Truncate if too long for the modal width
	maxEditTextWidth := modalWidth - 4 // Account for padding
	displayBuffer := model.TruncateWithEllipsis(editBuffer, maxEditTextWidth-1) + "|" // Add cursor
	editLine := editAreaStyle.Width(maxEditTextWidth).Render(displayBuffer)


	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Italic(true)
	helpLine := helpStyle.Render("Enter: Save | Esc: Cancel")

	// Combine modal content
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleLine,
		"", // Spacer
		editLine,
		"", // Spacer
		helpLine,
	)

	// Render the modal box with content
	return modalStyle.Render(content)
}