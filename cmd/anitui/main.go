package main

import (
	"fmt"
	"os"

	"github.com/anitui/anitui/internal/scraper"
	"github.com/anitui/anitui/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if os.Getenv("ANITUI_DEBUG") != "" {
		f, err := tea.LogToFile("/tmp/anitui-debug.log", "anitui")
		if err != nil {
			fmt.Fprintf(os.Stderr, "debug log setup: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	allanime := scraper.NewAllanimeScraper()
	unified := scraper.NewUnifiedScraper(allanime)

	model := tui.NewModel(unified)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "AniTUI error: %v\n", err)
		os.Exit(1)
	}
}
