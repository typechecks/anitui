package tui

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func modKey() string {
	if runtime.GOOS == "darwin" {
		return "⌘"
	}
	return "Ctrl"
}

func (m Model) View() string {
	switch m.screen {
	case ScreenHome:
		return m.viewHome()
	case ScreenSearching:
		return m.viewSearching()
	case ScreenResults:
		return m.viewResults()
	case ScreenEpisodes:
		return m.viewEpisodes()
	case ScreenLoadingEpisode:
		return m.viewLoadingEpisode()
	default:
		return m.viewHome()
	}
}

func (m Model) viewHome() string {
	var center strings.Builder

	logo := RenderAniTuiLogo(m.width)

	center.WriteString(logo)
	center.WriteString("\n\n")

	mode := "SUB"
	if m.dub {
		mode = "DUB"
	}
	modeLabel := lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true).
		Render(fmt.Sprintf(" [%s] ", mode))

	modeLine := lipgloss.JoinHorizontal(lipgloss.Center, modeLabel)
	center.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, modeLine))
	center.WriteString("\n\n")

	searchBox := SearchInputStyle.Width(min(60, m.width-10)).Render(m.input.View())
	searchBox = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, searchBox)
	center.WriteString(searchBox)
	center.WriteString("\n\n")

	help := HelpStyle.Render("Enter to search  |  Tab to switch " + mode + "  |  " + modKey() + "+C to quit")
	center.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help))

	version := lipgloss.NewStyle().
		Faint(true).
		Foreground(DimColor).
		Render("v" + Version)

	if m.height == 0 {
		return center.String() + "\n\n" + lipgloss.PlaceHorizontal(m.width, lipgloss.Right, version)
	}

	centered := lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, center.String())
	versionLine := lipgloss.PlaceHorizontal(m.width, lipgloss.Right, version)
	return centered + "\n" + versionLine
}

func (m Model) viewSearching() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Render("AniTUI")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	spinnerChar := spinnerChars[m.spinIndex%len(spinnerChars)]

	loading := fmt.Sprintf("%s %s", spinnerChar, m.loadingText)
	sb.WriteString(LoadingStyle.Render(loading))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, sb.String())
}

func (m Model) viewResults() string {
	if m.errorMsg != "" {
		return m.viewError(m.errorMsg)
	}

	var sb strings.Builder

	mode := "SUB"
	if m.dub {
		mode = "DUB"
	}
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		PaddingBottom(1).
		Render(fmt.Sprintf("Results for '%s' [%s] (%d found)", m.query, mode, len(m.results)))
	sb.WriteString(header)
	sb.WriteString("\n\n")

	availableHeight := m.height - 8
	startIdx := 0
	if m.cursor >= availableHeight {
		startIdx = m.cursor - availableHeight + 1
	}
	endIdx := min(len(m.results), startIdx+availableHeight)

	for i := startIdx; i < endIdx; i++ {
		anime := m.results[i]

		if i == m.cursor {
			title := SelectedListItemStyle.Width(m.width - 4).Render(anime.Title)
			sb.WriteString(title)
			sb.WriteString("\n")
			if anime.Description != "" {
				desc := lipgloss.NewStyle().
					Foreground(DimColor).
					Background(lipgloss.AdaptiveColor{Light: "7", Dark: "236"}).
					PaddingLeft(2).
					Width(m.width - 4).
					Render(truncate(anime.Description, m.width-8))
				sb.WriteString(desc)
				sb.WriteString("\n\n")
			}
		} else {
			title := ListItemStyle.Width(m.width - 4).Render(TitleStyle.Render(anime.Title))
			sb.WriteString(title)
			sb.WriteString("\n")
			if anime.Description != "" {
				desc := ListItemStyle.Width(m.width - 4).Render(DimStyle.Render(truncate(anime.Description, m.width-8)))
				sb.WriteString(desc)
				sb.WriteString("\n\n")
			}
		}
	}

	help := fmt.Sprintf("%d/%d results  |  ↑↓/jk navigate  |  Enter select  |  Esc back  |  / search", m.cursor+1, len(m.results))
	sb.WriteString(HelpStyle.Render(help))

	return sb.String()
}

func (m Model) viewEpisodes() string {
	if m.errorMsg != "" {
		return m.viewError(m.errorMsg)
	}

	var sb strings.Builder

	title := ""
	if m.selectedAnime != nil {
		title = m.selectedAnime.Title
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		PaddingBottom(1).
		Render(fmt.Sprintf("%s - Episodes (%d)", title, len(m.episodes)))
	sb.WriteString(header)
	sb.WriteString("\n\n")

	availableHeight := m.height - 8
	startIdx := 0
	if m.episodeCursor >= availableHeight {
		startIdx = m.episodeCursor - availableHeight + 1
	}
	endIdx := min(len(m.episodes), startIdx+availableHeight)

	for i := startIdx; i < endIdx; i++ {
		ep := m.episodes[i]
		prefix := "  "
		if i == m.episodeCursor {
			prefix = "▸ "
		}

		line := fmt.Sprintf("%s%s - %s", prefix, ep.Number, ep.Title)
		if i == m.episodeCursor {
			line = SelectedStyle.Render(line)
		} else {
			line = EpisodeStyle.Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	help := fmt.Sprintf("%d/%d episodes  |  ↑↓/jk navigate  |  Enter play  |  Esc back", m.episodeCursor+1, len(m.episodes))
	sb.WriteString("\n")
	sb.WriteString(HelpStyle.Render(help))

	return sb.String()
}

func (m Model) viewLoadingEpisode() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Render("AniTUI")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	spinnerChar := spinnerChars[m.spinIndex%len(spinnerChars)]

	loading := fmt.Sprintf("%s %s", spinnerChar, m.loadingText)
	sb.WriteString(LoadingStyle.Render(loading))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, sb.String())
}

func (m Model) viewError(errMsg string) string {
	var sb strings.Builder

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Render("AniTUI")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	errorBox := ErrorBoxStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ErrorColor).
		Padding(1, 2).
		Width(min(80, m.width-10)).
		Render(errMsg)

	errorBox = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, errorBox)
	sb.WriteString(errorBox)
	sb.WriteString("\n\n")

	help := HelpStyle.Render("Esc to go back  |  / to search again  |  " + modKey() + "+C to quit")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help))

	return lipgloss.PlaceVertical(m.height, lipgloss.Center, sb.String())
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
