package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const asciiArt = `
╔═══════════════════════════════════════════════════╗
║                                                   ║
║      █████╗ ███╗  ██╗██╗████████╗██╗   ██╗██╗     ║
║     ██╔══██╗████╗ ██║██║╚══██╔══╝██║   ██║██║     ║
║     ███████║██╔██╗██║██║   ██║   ██║   ██║██║     ║
║     ██╔══██║██║╚████║██║   ██║   ██║   ██║██║     ║
║     ██║  ██║██║ ╚███║██║   ██║   ╚██████╔╝██║     ║
║     ╚═╝  ╚═╝╚═╝  ╚══╝╚═╝   ╚═╝    ╚═════╝ ╚═╝     ║
║                                                   ║
║          Watch anime from your terminal           ║
╚═══════════════════════════════════════════════════╝`

var (
	accent    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	secondary = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
)

func RenderAniTuiLogo(termWidth int) string {
	if termWidth <= 0 {
		termWidth = 80
	}

	lines := strings.Split(strings.TrimSpace(asciiArt), "\n")
	var b strings.Builder

	splitPoint := 26

	for _, line := range lines {
		runes := []rune(line)
		for i, ch := range runes {
			isLogoLetter := (ch >= '█' && ch <= '')

			if isLogoLetter {
				if i < splitPoint {
					b.WriteString(accent.Render(string(ch)))
				} else {
					b.WriteString(secondary.Render(string(ch)))
				}
			} else {
				b.WriteRune(ch)
			}
		}
		b.WriteRune('\n')
	}

	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, b.String())
}
