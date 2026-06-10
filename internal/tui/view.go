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
	return "ctrl"
}

func (m Model) View() string {
	switch m.screen {
	case ScreenHome:
		return m.viewHome()
	case ScreenSearching:
		return m.viewLoading(m.loadingText)
	case ScreenResults:
		return m.viewResults()
	case ScreenEpisodes:
		return m.viewEpisodes()
	case ScreenLoadingEpisode:
		return m.viewLoading(m.loadingText)
	case ScreenWatching:
		return m.viewWatching()
	default:
		return m.viewHome()
	}
}

func (m Model) viewHome() string {
	var sb strings.Builder

	logo := RenderAniTuiLogo(m.width)
	sb.WriteString(logo)
	sb.WriteString("\n\n")

	searchBox := SearchInputStyle.Width(min(60, m.width-10)).Render(m.input.View())
	searchBox = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, searchBox)
	sb.WriteString(searchBox)
	sb.WriteString("\n\n")

	// Centered search hint
	hintStyle := lipgloss.NewStyle().
		Foreground(DimColor).
		Faint(true)
	hintText := "Search for an anime..."
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, hintStyle.Render(hintText)))
	sb.WriteString("\n")

	contentStr := sb.String()

	versionStr := "v" + Version
	helpText := "enter [search] |  " + modKey() + "+c [quit]"
	helpWidth := len([]rune(helpText))
	versionWidth := len([]rune(versionStr))
	gaps := m.width - helpWidth - versionWidth
	if gaps < 1 {
		gaps = 1
	}
	helpLine := HelpStyle.Render(helpText + strings.Repeat(" ", gaps) + versionStr)

	separatorLine := DimStyle.Render(strings.Repeat("─", m.width))

	if m.height == 0 {
		return contentStr + "\n\n" + separatorLine + "\n" + helpLine
	}

	// Footer is 3 lines: separator + blank(PaddingTop) + help text
	footerHeight := 3
	contentHeight := strings.Count(contentStr, "\n") + 1
	available := m.height - footerHeight
	if available < contentHeight {
		available = contentHeight
	}
	topPad := (available - contentHeight) / 2
	bottomPad := available - contentHeight - topPad

	var result strings.Builder
	result.WriteString(strings.Repeat("\n", topPad))
	result.WriteString(contentStr)
	result.WriteString(strings.Repeat("\n", bottomPad))
	result.WriteString(separatorLine)
	result.WriteString("\n")
	result.WriteString(helpLine)
	return result.String()
}

func (m Model) viewLoading(text string) string {
	var sb strings.Builder

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Render("AniTUI")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	spinnerChar := spinnerChars[m.spinIndex%len(spinnerChars)]
	loading := fmt.Sprintf("%s %s", spinnerChar, text)
	sb.WriteString(LoadingStyle.Render(loading))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, sb.String())
}

func (m Model) viewResults() string {
	if m.errorMsg != "" {
		return m.viewError(m.errorMsg)
	}

	var sb strings.Builder

	header := TitleStyle.
		PaddingBottom(1).
		Render(fmt.Sprintf("Results for '%s' (%d found)", m.query, len(m.results)))
	sb.WriteString(header)
	sb.WriteString("\n\n")

	maxLines := max(1, m.height-8)

	// Scroll backward from cursor to find the starting index.
	// Unselected items take 2 lines (title + metadata); the selected item
	// takes up to 3 lines (title + metadata + synopsis).
	startIdx := m.cursor
	linesBefore := 0
	for startIdx > 0 {
		if linesBefore+2 >= maxLines {
			break
		}
		linesBefore += 2
		startIdx--
	}

	// Pre-create styles for metadata and synopsis lines.
	metaStyle := lipgloss.NewStyle().
		Foreground(DimColor).
		PaddingLeft(2)
	synopsisStyle := lipgloss.NewStyle().
		Foreground(DimColor).
		Faint(true).
		PaddingLeft(2)
	selBgColor := lipgloss.AdaptiveColor{Light: "7", Dark: "236"}

	linesUsed := 0
	for i := startIdx; i < len(m.results); i++ {
		anime := m.results[i]

		// ---- Title line ----
		if i == m.cursor {
			sb.WriteString(SelectedListItemStyle.Render("▸ " + anime.Title))
			sb.WriteString("\n")
		} else {
			sb.WriteString(ListItemStyle.Render(TitleStyle.Render(anime.Title)))
			sb.WriteString("\n")
		}
		linesUsed++

		// ---- Metadata line ----
		var metaParts []string
		animeType := anime.Type
		if animeType == "" {
			animeType = "?"
		}
		metaParts = append(metaParts, fmt.Sprintf("[%s]", animeType))
		if anime.Year > 0 {
			metaParts = append(metaParts, fmt.Sprintf("%d", anime.Year))
		}
		if anime.EpisodeCount > 0 {
			metaParts = append(metaParts, fmt.Sprintf("%d eps", anime.EpisodeCount))
		}
		if len(metaParts) > 0 {
			metaLine := strings.Join(metaParts, "  |  ")
			if i == m.cursor {
				sb.WriteString(metaStyle.Background(selBgColor).Render(metaLine))
			} else {
				sb.WriteString(metaStyle.Render(metaLine))
			}
		}
		sb.WriteString("\n")
		linesUsed++

		// ---- Synopsis line (selected item only) ----
		if i == m.cursor {
			synopsis := anime.Synopsis
			if synopsis == "" {
				synopsis = anime.Description
			}
			if synopsis != "" {
				if i == m.cursor {
					sb.WriteString(synopsisStyle.Background(selBgColor).Render(truncate(synopsis, m.width-6)))
				} else {
					sb.WriteString(synopsisStyle.Render(truncate(synopsis, m.width-6)))
				}
				sb.WriteString("\n")
				linesUsed++
			}
		}

		if linesUsed >= maxLines {
			break
		}
	}

	contentStr := sb.String()

	versionStr := "v" + Version
	help := fmt.Sprintf("%d results  |  ↑↓/jk [navigate]  |  gg/G [page up/down]  |  enter [select]  |  esc [back]  |  / [search]",
		len(m.results))
	helpWidth := len([]rune(help))
	versionWidth := len([]rune(versionStr))
	gaps := m.width - helpWidth - versionWidth
	if gaps < 1 {
		gaps = 1
	}
	helpLine := HelpStyle.Render(help + strings.Repeat(" ", gaps) + versionStr)

	separatorLine := DimStyle.Render(strings.Repeat("─", m.width))

	if m.height == 0 {
		return contentStr + separatorLine + "\n" + helpLine
	}

	footerHeight := 3 // separator + blank(PaddingTop) + help
	contentHeight := strings.Count(contentStr, "\n") + 1

	var result strings.Builder
	result.WriteString(contentStr)
	if contentHeight+footerHeight < m.height {
		result.WriteString(strings.Repeat("\n", m.height-contentHeight-footerHeight))
	}
	result.WriteString(separatorLine)
	result.WriteString("\n")
	result.WriteString(helpLine)
	return result.String()
}

func (m Model) viewEpisodes() string {
	if m.errorMsg != "" {
		return m.viewError(m.errorMsg)
	}

	var sb strings.Builder

	title := ""
	anime := m.selectedAnime
	if anime != nil {
		title = anime.Title
	}

	// ---- Title line ----
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Render(fmt.Sprintf("%s — Episodes (%d)", title, len(m.episodes)))
	sb.WriteString(header)

	// ---- Metadata line ----
	if anime != nil {
		var metaParts []string

		// [Type]
		if anime.Type != "" {
			metaParts = append(metaParts, fmt.Sprintf("[%s]", anime.Type))
		}

		// Year
		if anime.Year > 0 {
			metaParts = append(metaParts, fmt.Sprintf("%d", anime.Year))
		}

		// ✪ Score/10
		if anime.Score > 0 {
			metaParts = append(metaParts, fmt.Sprintf("✪ %.1f/10", anime.Score))
		}

		// Studio
		if anime.Studio != "" {
			metaParts = append(metaParts, anime.Studio)
		}

		// Genres
		if len(anime.Genres) > 0 {
			metaParts = append(metaParts, strings.Join(anime.Genres, ", "))
		}

		// Status (formatted nicely)
		status := formatStatus(anime.Status)
		if status != "" {
			metaParts = append(metaParts, status)
		}

		if len(metaParts) > 0 {
			metaStyle := lipgloss.NewStyle().
				Foreground(DimColor).
				PaddingLeft(2)
			metaLine := metaStyle.Render(strings.Join(metaParts, "  |  "))
			sb.WriteString("\n")
			sb.WriteString(metaLine)
		}
	}

	// ---- Synopsis section (collapsible with space) ----
	if anime != nil && anime.Synopsis != "" {
		sb.WriteString("\n\n")
		synopsis := anime.Synopsis
		synopsisStyle := lipgloss.NewStyle().
			Foreground(DimColor).
			Faint(true).
			PaddingLeft(2)
		if m.showFullSynopsis {
			// Show full synopsis
			wrapped := lipgloss.NewStyle().Width(m.width - 4).Render(synopsis)
			sb.WriteString(synopsisStyle.Render(wrapped))
			sb.WriteString("\n")
			sb.WriteString(DimStyle.Render("[space to collapse]"))
		} else {
			// Show first ~4 lines
			wrapped := lipgloss.NewStyle().Width(m.width - 4).Render(synopsis)
			lines := strings.Split(wrapped, "\n")
			showLines := min(4, len(lines))
			sb.WriteString(synopsisStyle.Render(strings.Join(lines[:showLines], "\n")))
			if len(lines) > 4 {
				sb.WriteString("\n")
				sb.WriteString(DimStyle.Render("[space to expand]"))
			}
		}
	}

	sb.WriteString("\n\n")

	// ---- Episode list ----
	availableHeight := max(1, m.height-10)
	if anime != nil && anime.Synopsis != "" {
		if m.showFullSynopsis {
			// More lines used for full synopsis — fewer available for episode list
			availableHeight = max(1, m.height-14)
		} else {
			availableHeight = max(1, m.height-12)
		}
	}
	if availableHeight < 1 {
		availableHeight = 1
	}

	startIdx := 0
	if m.episodeCursor >= availableHeight {
		startIdx = max(0, m.episodeCursor-availableHeight+1)
	}
	endIdx := min(len(m.episodes), startIdx+availableHeight)

	for i := startIdx; i < endIdx; i++ {
		ep := m.episodes[i]
		prefix := "  "
		if i == m.episodeCursor {
			prefix = "▸ "
		}

		epNum := strings.TrimPrefix(ep.Number, "EP ")
		epNum = strings.TrimPrefix(epNum, "Episode ")
		line := fmt.Sprintf("%sEP %s - %s", prefix, epNum, ep.Title)
		if i == m.episodeCursor {
			line = SelectedStyle.Render(line)
		} else {
			line = EpisodeStyle.Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	contentStr := sb.String()

	// ---- Help bar ----
	versionStr := "v" + Version
	help := fmt.Sprintf("%d/%d episodes  |  ↑↓/jk [nav]  |  enter [play]  |  esc [back]",
		m.episodeCursor+1, len(m.episodes))
	helpWidth := len([]rune(help))
	versionWidth := len([]rune(versionStr))
	gaps := m.width - helpWidth - versionWidth
	if gaps < 1 {
		gaps = 1
	}
	helpLine := HelpStyle.Render(help + strings.Repeat(" ", gaps) + versionStr)

	separatorLine := DimStyle.Render(strings.Repeat("─", m.width))

	if m.height == 0 {
		return contentStr + separatorLine + "\n" + helpLine
	}

	footerHeight := 3
	contentHeight := strings.Count(contentStr, "\n") + 1

	var result strings.Builder
	result.WriteString(contentStr)
	if contentHeight+footerHeight < m.height {
		result.WriteString(strings.Repeat("\n", m.height-contentHeight-footerHeight))
	}
	result.WriteString(separatorLine)
	result.WriteString("\n")
	result.WriteString(helpLine)
	return result.String()
}

func (m Model) viewWatching() string {
	if m.watching == nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, LoadingStyle.Render("Loading..."))
	}

	var sb strings.Builder

	// Anime title
	title := m.watching.animeTitle
	if title == "" && m.selectedAnime != nil {
		title = m.selectedAnime.Title
	}
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, TitleStyle.Render(title)))
	sb.WriteString("\n")

	// Episode line
	var epNumber, epTitle string
	if m.episodeCursor >= 0 && m.episodeCursor < len(m.episodes) {
		epNumber = m.episodes[m.episodeCursor].Number
		epTitle = m.episodes[m.episodeCursor].Title
	}
	sb.WriteString("\n")
	epNum := strings.TrimPrefix(epNumber, "EP ")
	epNum = strings.TrimPrefix(epNum, "Episode ")
	epLine := fmt.Sprintf("EP %s - %s", epNum, epTitle)
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, epLine))
	sb.WriteString("\n")

	// Status
	sb.WriteString("\n")
	statusLine := LoadingStyle.Render("Now playing in external player...")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, statusLine))
	sb.WriteString("\n")

	// Error message (if any) — shown above controls
	if m.errorMsg != "" {
		sb.WriteString("\n")
		errBox := ErrorBoxStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ErrorColor).
			Padding(1, 2).
			Width(min(80, m.width-10)).
			Render(m.errorMsg)
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, errBox))
		sb.WriteString("\n")
	}

	// Controls
	sb.WriteString("\n")
	var btnControls []string
	if m.watching.episodeIndex > 0 {
		btnControls = append(btnControls, ControlStyle.Render("← prev"))
	}
	btnControls = append(btnControls, ControlStyle.Render("space [replay]"))
	if m.watching.episodeIndex < m.watching.episodesLen-1 {
		btnControls = append(btnControls, ControlStyle.Render("next →"))
	}
	controlsStr := strings.Join(btnControls, "  |  ")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, controlsStr))
	sb.WriteString("\n")

	// Source line
	quality := ""
	sourceCount := 0
	if m.watching != nil {
		sourceCount = len(m.watching.sources)
		if m.watching.sourceIndex >= 0 && m.watching.sourceIndex < len(m.watching.sources) {
			quality = m.watching.sources[m.watching.sourceIndex].Quality
		}
	}
	mode := "SUB"
	if m.watching.dub {
		mode = "DUB"
	}
	sourcePart := ""
	if sourceCount > 1 {
		sourcePart = fmt.Sprintf("Source %d/%d", m.watching.sourceIndex+1, sourceCount)
		if quality != "" && quality != "default" {
			sourcePart += fmt.Sprintf(" — %s", quality)
		}
	} else if quality != "" && quality != "default" {
		sourcePart = fmt.Sprintf("Source: %s", quality)
	} else {
		sourcePart = "Source:"
	}
	sourceLine := DimStyle.Render(fmt.Sprintf("%s  |  %s", sourcePart, mode))
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, sourceLine))
	sb.WriteString("\n")

	contentStr := sb.String()

	versionStr := "v" + Version
	helpText := "⇆/hl [prev/next]  |  space [replay]  |  s [source]  |  d [sub/dub]  |  esc [back]"
	helpWidth := len([]rune(helpText))
	versionWidth := len([]rune(versionStr))
	gaps := m.width - helpWidth - versionWidth
	if gaps < 1 {
		gaps = 1
	}
	helpLine := HelpStyle.Render(helpText + strings.Repeat(" ", gaps) + versionStr)

	separatorLine := DimStyle.Render(strings.Repeat("─", m.width))

	if m.height == 0 {
		return contentStr + separatorLine + "\n" + helpLine
	}

	footerHeight := 3
	contentHeight := strings.Count(contentStr, "\n") + 1
	availableHeight := m.height - footerHeight
	if availableHeight < contentHeight {
		availableHeight = contentHeight
	}
	topPad := (availableHeight - contentHeight) / 2
	bottomPad := availableHeight - contentHeight - topPad

	var result strings.Builder
	result.WriteString(strings.Repeat("\n", topPad))
	result.WriteString(contentStr)
	result.WriteString(strings.Repeat("\n", bottomPad))
	result.WriteString(separatorLine)
	result.WriteString("\n")
	result.WriteString(helpLine)
	return result.String()
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
	sb.WriteString("\n")

	contentStr := sb.String()

	versionStr := "v" + Version
	help := "esc [back]  |  / [search]  |  " + modKey() + "+c [quit]"
	helpWidth := len([]rune(help))
	versionWidth := len([]rune(versionStr))
	gaps := m.width - helpWidth - versionWidth
	if gaps < 1 {
		gaps = 1
	}
	helpLine := HelpStyle.Render(help + strings.Repeat(" ", gaps) + versionStr)

	separatorLine := DimStyle.Render(strings.Repeat("─", m.width))

	if m.height == 0 {
		return contentStr + separatorLine + "\n" + helpLine
	}

	footerHeight := 3 // separator + blank(PaddingTop) + help
	contentHeight := strings.Count(contentStr, "\n") + 1

	var result strings.Builder
	result.WriteString(contentStr)
	if contentHeight+footerHeight < m.height {
		result.WriteString(strings.Repeat("\n", m.height-contentHeight-footerHeight))
	}
	result.WriteString(separatorLine)
	result.WriteString("\n")
	result.WriteString(helpLine)
	return result.String()
}

func truncate(s string, maxLen int) string {
	// Collapse all whitespace (newlines, multiple spaces) to single spaces
	s = strings.Join(strings.Fields(s), " ")

	s = strings.TrimSpace(s)

	// Use rune count for proper Unicode handling
	runes := []rune(s)
	if len(runes) <= maxLen {
		return string(runes)
	}

	return string(runes[:maxLen-3]) + "..."
}

func formatStatus(status string) string {
	switch status {
	case "FINISHED":
		return "Finished Airing"
	case "RELEASING":
		return "Currently Airing"
	case "NOT_YET_RELEASED":
		return "Not Yet Aired"
	case "CANCELLED":
		return "Cancelled"
	case "HIATUS":
		return "On Hiatus"
	default:
		return status
	}
}
