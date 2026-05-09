package tui

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/anitui/anitui/internal/models"
	"github.com/anitui/anitui/internal/player"
	"github.com/anitui/anitui/internal/scraper"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Screen int

const (
	ScreenHome Screen = iota
	ScreenSearching
	ScreenResults
	ScreenEpisodes
	ScreenLoadingEpisode
)

var Version = "dev"

var debugLog *log.Logger

func init() {
	if os.Getenv("ANITUI_DEBUG") != "" {
		logPath := "anitui-tui-debug.log"
		if tempDir := os.TempDir(); tempDir != "" {
			logPath = os.ExpandEnv(fmt.Sprintf("%s/anitui-tui-debug.log", tempDir))
		}
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			debugLog = log.New(io.Discard, "", 0)
			return
		}
		debugLog = log.New(f, "", log.LstdFlags)
	} else {
		debugLog = log.New(io.Discard, "", 0)
	}
}

type searchResultsMsg struct {
	results []models.Anime
	err     error
}

type episodesMsg struct {
	episodes []models.Episode
	err      error
}

type videoSourcesMsg struct {
	sources []models.VideoSource
	err     error
}

type playDoneMsg struct {
	err error
}

type tickMsg time.Time

type Model struct {
	screen    Screen
	input     textinput.Model
	width     int
	height    int

	scrapers      *scraper.UnifiedScraper
	currentSource scraper.Scraper

	query       string
	results     []models.Anime
	cursor      int
	errorMsg    string
	loadingText string

	episodes      []models.Episode
	selectedAnime *models.Anime
	episodeCursor int

	spinIndex int

	lastKey string
	dub     bool

	loadingSince time.Time
	pendingMsg   tea.Msg
	minLoadTime  time.Duration
}

func NewModel(scrapers *scraper.UnifiedScraper) Model {
	ti := textinput.New()
	ti.Placeholder = "Search anime..."
	ti.CharLimit = 100
	ti.Width = 60
	ti.Prompt = "> "
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "12"})
	ti.Focus()

	return Model{
		screen:       ScreenHome,
		input:        ti,
		scrapers:     scrapers,
		dub:          false,
		minLoadTime:  600 * time.Millisecond,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case searchResultsMsg:
		if time.Since(m.loadingSince) < m.minLoadTime {
			m.pendingMsg = msg
			return m, nil
		}
		m.pendingMsg = nil
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			m.screen = ScreenResults
			return m, nil
		}
		m.results = msg.results
		m.cursor = 0
		m.screen = ScreenResults
		m.errorMsg = ""
		if len(m.results) == 0 {
			m.errorMsg = fmt.Sprintf("No results found for '%s'", m.query)
		}
		return m, nil

	case episodesMsg:
		if time.Since(m.loadingSince) < m.minLoadTime {
			m.pendingMsg = msg
			return m, nil
		}
		m.pendingMsg = nil
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			m.screen = ScreenEpisodes
			return m, nil
		}
		m.episodes = msg.episodes
		m.episodeCursor = 0
		m.screen = ScreenEpisodes
		m.errorMsg = ""
		return m, nil

	case videoSourcesMsg:
		debugLog.Printf("[TUI] videoSourcesMsg: err=%v sources=%d screen=%d", msg.err, len(msg.sources), m.screen)
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to load video: %v", msg.err)
			m.screen = ScreenEpisodes
			m.pendingMsg = nil
			return m, nil
		}
		if len(msg.sources) == 0 {
			m.errorMsg = "No video sources found for this episode"
			m.screen = ScreenEpisodes
			m.pendingMsg = nil
			return m, nil
		}

		if time.Since(m.loadingSince) < m.minLoadTime {
			m.pendingMsg = msg
			return m, nil
		}

		debugLog.Printf("[TUI] Got %d sources, launching player. URL=%s", len(msg.sources), msg.sources[0].URL)
		m.screen = ScreenEpisodes
		m.pendingMsg = nil
		return m, m.playEpisode(msg.sources)

	case playDoneMsg:
		debugLog.Printf("[TUI] playDoneMsg: err=%v", msg.err)
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Player error: %v", msg.err)
		}
		return m, nil

	case tickMsg:
		m.spinIndex++
		if m.pendingMsg != nil && time.Since(m.loadingSince) >= m.minLoadTime {
			p := m.pendingMsg
			m.pendingMsg = nil
			switch msg := p.(type) {
			case searchResultsMsg:
				if msg.err != nil {
					m.errorMsg = msg.err.Error()
				} else {
					m.results = msg.results
					m.cursor = 0
					m.errorMsg = ""
					if len(m.results) == 0 {
						m.errorMsg = fmt.Sprintf("No results found for '%s'", m.query)
					}
				}
				m.screen = ScreenResults
				return m, nil
			case episodesMsg:
				if msg.err != nil {
					m.errorMsg = msg.err.Error()
				} else {
					m.episodes = msg.episodes
					m.episodeCursor = 0
					m.errorMsg = ""
				}
				m.screen = ScreenEpisodes
				return m, nil
			case videoSourcesMsg:
				if msg.err == nil && len(msg.sources) > 0 {
					debugLog.Printf("[TUI] Delayed play: URL=%s", msg.sources[0].URL)
					m.screen = ScreenEpisodes
					return m, m.playEpisode(msg.sources)
				}
				m.screen = ScreenEpisodes
				return m, nil
			}
		}
		if m.screen == ScreenSearching || m.screen == ScreenLoadingEpisode {
			return m, tickCmd()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.screen == ScreenHome {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "/":
			m.input.Focus()
			return m, textinput.Blink
		case "tab":
			m.dub = !m.dub
			if m.dub {
				m.input.Placeholder = "Search anime (dub)..."
			} else {
				m.input.Placeholder = "Search anime..."
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	key := msg.String()
	isDoubleG := key == "g" && m.lastKey == "g"
	m.lastKey = key

	switch {
	case key == "ctrl+c":
		return m, tea.Quit

	case key == "esc":
		return m.handleEsc()

	case key == "enter":
		return m.handleEnter()

	case key == "up" || key == "k":
		return m.handleUp()

	case key == "down" || key == "j":
		return m.handleDown()

	case key == "g" && isDoubleG:
		return m.handleG()

	case key == "G":
		return m.handleCapitalG()

	case key == "ctrl+u":
		return m.handleCtrlU()

	case key == "ctrl+d":
		return m.handleCtrlD()

	case key == "/":
		return m.handleSlash()
	}

	return m, nil
}

func (m Model) handleEsc() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenResults:
		m.screen = ScreenHome
		m.results = nil
		m.cursor = 0
		m.errorMsg = ""
		m.input.Reset()
		m.input.Focus()
		return m, textinput.Blink
	case ScreenEpisodes:
		m.screen = ScreenResults
		m.episodes = nil
		m.episodeCursor = 0
		m.errorMsg = ""
		m.cursor = 0
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenHome:
		query := strings.TrimSpace(m.input.Value())
		if query == "" {
			return m, nil
		}
		m.query = query
		m.screen = ScreenSearching
		m.loadingText = fmt.Sprintf("Searching for '%s'...", query)
		m.spinIndex = 0
		m.errorMsg = ""
		m.loadingSince = time.Now()
		return m, tea.Batch(m.performSearch(query), tickCmd())

	case ScreenResults:
		if len(m.results) == 0 {
			return m, nil
		}
		if m.cursor >= len(m.results) {
			m.cursor = 0
		}
		anime := m.results[m.cursor]
		m.selectedAnime = &anime
		m.screen = ScreenLoadingEpisode
		m.loadingText = fmt.Sprintf("Loading episodes for '%s'...", anime.Title)
		m.spinIndex = 0
		m.errorMsg = ""
		m.loadingSince = time.Now()
		m.pendingMsg = nil

		m.currentSource = m.scrapers.GetScraper(anime.Source)
		if m.currentSource == nil {
			// Fallback or handle error
			m.currentSource = m.scrapers.GetScraper("allanime.day")
		}
		return m, tea.Batch(m.loadEpisodes(anime.URL), tickCmd())

	case ScreenEpisodes:
		if len(m.episodes) == 0 {
			return m, nil
		}
		if m.episodeCursor >= len(m.episodes) {
			m.episodeCursor = 0
		}
		episode := m.episodes[m.episodeCursor]
		m.screen = ScreenLoadingEpisode
		m.loadingText = fmt.Sprintf("Fetching video for %s...", episode.Title)
		m.spinIndex = 0
		m.errorMsg = ""
		m.loadingSince = time.Now()
		m.pendingMsg = nil
		return m, tea.Batch(m.loadVideoURL(episode.URL), tickCmd())
	}

	return m, nil
}

func (m Model) handleUp() (tea.Model, tea.Cmd) {
	if m.screen == ScreenHome {
		return m, nil
	}
	switch m.screen {
	case ScreenResults:
		if m.cursor > 0 {
			m.cursor--
		}
	case ScreenEpisodes:
		if m.episodeCursor > 0 {
			m.episodeCursor--
		}
	}
	return m, nil
}

func (m Model) handleDown() (tea.Model, tea.Cmd) {
	if m.screen == ScreenHome {
		return m, nil
	}
	switch m.screen {
	case ScreenResults:
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
	case ScreenEpisodes:
		if m.episodeCursor < len(m.episodes)-1 {
			m.episodeCursor++
		}
	}
	return m, nil
}

func (m Model) handleG() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenResults:
		if len(m.results) > 0 {
			m.cursor = 0
		}
	case ScreenEpisodes:
		if len(m.episodes) > 0 {
			m.episodeCursor = 0
		}
	}
	return m, nil
}

func (m Model) handleCapitalG() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenResults:
		if len(m.results) > 0 {
			m.cursor = len(m.results) - 1
		}
	case ScreenEpisodes:
		if len(m.episodes) > 0 {
			m.episodeCursor = len(m.episodes) - 1
		}
	}
	return m, nil
}

func (m Model) handleCtrlU() (tea.Model, tea.Cmd) {
	pageSize := 10
	switch m.screen {
	case ScreenResults:
		m.cursor = max(0, m.cursor-pageSize)
	case ScreenEpisodes:
		m.episodeCursor = max(0, m.episodeCursor-pageSize)
	}
	return m, nil
}

func (m Model) handleCtrlD() (tea.Model, tea.Cmd) {
	pageSize := 10
	switch m.screen {
	case ScreenResults:
		if len(m.results) > 0 {
			m.cursor = min(len(m.results)-1, m.cursor+pageSize)
		}
	case ScreenEpisodes:
		if len(m.episodes) > 0 {
			m.episodeCursor = min(len(m.episodes)-1, m.episodeCursor+pageSize)
		}
	}
	return m, nil
}

func (m Model) handleSlash() (tea.Model, tea.Cmd) {
	if m.screen != ScreenHome {
		m.screen = ScreenHome
		m.results = nil
		m.cursor = 0
		m.errorMsg = ""
		m.input.Reset()
		m.input.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) performSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results := m.scrapers.Search(query, m.dub)
		return searchResultsMsg{results: results}
	}
}

func (m Model) loadEpisodes(url string) tea.Cmd {
	return func() tea.Msg {
		episodes, err := m.currentSource.GetEpisodes(url, m.dub)
		return episodesMsg{episodes: episodes, err: err}
	}
}

func (m Model) loadVideoURL(url string) tea.Cmd {
	return func() tea.Msg {
		sources, err := m.currentSource.GetVideoURL(url, m.dub)
		return videoSourcesMsg{sources: sources, err: err}
	}
}

func (m Model) playEpisode(sources []models.VideoSource) tea.Cmd {
	return func() tea.Msg {
		videoURL := sources[0].URL
		debugLog.Printf("[TUI] playEpisode: calling player.Play(%s)", videoURL)
		err := player.Play(videoURL)
		debugLog.Printf("[TUI] playEpisode: player.Play returned err=%v", err)
		return playDoneMsg{err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
