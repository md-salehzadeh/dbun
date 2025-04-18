package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	tables         []string // List of tables
	selectedIdx    int      // Currently selected table index
	activeTableIdx int      // Currently active table index
	mode           string   // Mode: "Data", "Structure", "Indices"
	width          int      // Terminal width
	height         int      // Terminal height
	focusLeft      bool     // Focus on left (true) or right (false) box
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

	// Main Box: Show selected table data based on mode
	mainContent := fmt.Sprintf("Showing %s for: %s", m.mode, m.tables[m.activeTableIdx])
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

func main() {
	m := model{
		tables:         []string{"users", "orders", "products", "categories"},
		selectedIdx:    0,
		activeTableIdx: 0, // Initialize active table index
		mode:           "Data", // Start with Data view
		width:          80,     // Default width
		height:         24,     // Default height
		focusLeft:      true,   // Start with focus on left box
	}

	p := tea.NewProgram(&m, tea.WithAltScreen()) // Fullscreen app

	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
