package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	wishbubble "github.com/charmbracelet/wish/bubbletea"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	accent    = lipgloss.Color("#7C3AED")
	subtle    = lipgloss.Color("#6B7280")
	highlight = lipgloss.Color("#F3F0FF")
	green     = lipgloss.Color("#10B981")
	amber     = lipgloss.Color("#F59E0B")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			MarginBottom(1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(accent).
			Width(60).
			MarginTop(1).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Background(accent).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB")).
			Padding(0, 1)

	dimStyle = lipgloss.NewStyle().
			Foreground(subtle)

	tagStyle = lipgloss.NewStyle().
			Foreground(accent).
			Background(lipgloss.Color("#1E1B4B")).
			Padding(0, 1).
			MarginRight(1)

	badgeStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			Underline(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 2).
			Width(60)
)

// ── Data ──────────────────────────────────────────────────────────────────────

type Project struct {
	Name   string
	Status string
	Desc   string
	Tech   []string
	URL    string
}

type Job struct {
	Title   string
	Company string
	Period  string
	Bullets []string
	Tech    []string
}

var projects = []Project{
	{
		Name:   "Findable",
		Status: "Live",
		Desc:   "Anonymous lost-item recovery system. Scan a QR code → send message → owner gets notified. Privacy-focused.",
		Tech:   []string{"Next.js", "PostgreSQL", "Email"},
		URL:    "findable.itzpankaj.site",
	},
	{
		Name:   "Motivation Kaksha",
		Status: "50k+ users",
		Desc:   "JEE college prediction platform. High-performance REST API with Redis caching layer.",
		Tech:   []string{"Express", "Redis", "Node.js"},
		URL:    "motivationkaksha.in",
	},
	{
		Name:   "Kaksha AI",
		Status: "Building",
		Desc:   "RAG-based educational assistant. Explains complex problems step-by-step using context-aware LLM agents.",
		Tech:   []string{"GenAI", "RAG", "Python"},
		URL:    "",
	},
	{
		Name:   "Book Reader",
		Status: "Live",
		Desc:   "Clean PDF reader app with smooth navigation and a minimal interface.",
		Tech:   []string{"Next.js", "TypeScript", "Tailwind"},
		URL:    "reader.itzpankaj.site",
	},
	{
		Name:   "Kat n Trim",
		Status: "Live",
		Desc:   "YouTube video trimmer & creator engagement tool with real-time webhook processing.",
		Tech:   []string{"ytdl", "Youtube API", "Node.js"},
		URL:    "v0-kat-ntrim-project.vercel.app",
	},
}

var jobs = []Job{
	{
		Title:   "Founding Engineer",
		Company: "Working (Startup)",
		Period:  "Apr 2024 – Present · Remote",
		Bullets: []string{
			"Migrated legacy codebase to Next.js → 40% perf boost",
			"Deployed on AWS EC2, built REST APIs with Django & FastAPI",
			"Implemented RAG-based AI system",
		},
		Tech: []string{"Next.js", "Django", "FastAPI", "AWS", "RAG", "PostgreSQL"},
	},
	{
		Title:   "Frontend Developer Intern",
		Company: "Startup",
		Period:  "Jan 2024 – Mar 2024 · Remote",
		Bullets: []string{
			"Built responsive UI components with Next.js & Tailwind CSS",
			"Translated Figma mockups into maintainable frontend code",
		},
		Tech: []string{"Next.js", "Tailwind CSS", "TypeScript", "Figma"},
	},
}

var skills = map[string][]string{
	"Languages": {"TypeScript", "JavaScript", "Python", "Rust", "Go", "C"},
	"Frontend":  {"React", "Next.js", "Tailwind CSS"},
	"Backend":   {"Node.js", "Django", "FastAPI", "Express"},
	"Infra":     {"PostgreSQL", "MongoDB", "Redis", "AWS", "GCP", "Docker"},
}

// ── Pages ─────────────────────────────────────────────────────────────────────

type page int

const (
	pageMenu page = iota
	pageAbout
	pageExperience
	pageProjects
	pageSkills
	pageContact
)

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	page        page
	cursor      int
	projectIdx  int
	width       int
	height      int
}

var menuItems = []string{
	"About",
	"Experience",
	"Projects",
	"Skills",
	"Contact",
	"Exit",
}

func initialModel() model {
	return model{page: pageMenu}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			if m.page == pageMenu {
				return m, tea.Quit
			}
			m.page = pageMenu
			m.cursor = 0

		case "esc":
			if m.page != pageMenu {
				m.page = pageMenu
				m.cursor = 0
			}

		case "up", "k":
			if m.page == pageMenu && m.cursor > 0 {
				m.cursor--
			} else if m.page == pageProjects && m.projectIdx > 0 {
				m.projectIdx--
			}

		case "down", "j":
			if m.page == pageMenu && m.cursor < len(menuItems)-1 {
				m.cursor++
			} else if m.page == pageProjects && m.projectIdx < len(projects)-1 {
				m.projectIdx++
			}

		case "enter", " ":
			if m.page == pageMenu {
				switch m.cursor {
				case 0:
					m.page = pageAbout
				case 1:
					m.page = pageExperience
				case 2:
					m.page = pageProjects
					m.projectIdx = 0
				case 3:
					m.page = pageSkills
				case 4:
					m.page = pageContact
				case 5:
					return m, tea.Quit
				}
				m.cursor = 0
			}
		}
	}

	return m, nil
}

// ── Views ─────────────────────────────────────────────────────────────────────

func (m model) View() string {
	switch m.page {
	case pageAbout:
		return m.viewAbout()
	case pageExperience:
		return m.viewExperience()
	case pageProjects:
		return m.viewProjects()
	case pageSkills:
		return m.viewSkills()
	case pageContact:
		return m.viewContact()
	default:
		return m.viewMenu()
	}
}

func (m model) viewMenu() string {
	var b strings.Builder

	banner := `
██████╗  █████╗ ███╗   ██╗██╗  ██╗ █████╗      ██╗ ██████╗ ███████╗
██╔══██╗██╔══██╗████╗  ██║██║ ██╔╝██╔══██╗     ██║██╔═══██╗██╔════╝
██████╔╝███████║██╔██╗ ██║█████╔╝ ███████║     ██║██║   ██║███████╗
██╔═══╝ ██╔══██║██║╚██╗██║██╔═██╗ ██╔══██║██   ██║██║   ██║╚════██║
██║     ██║  ██║██║ ╚████║██║  ██╗██║  ██║╚█████╔╝╚██████╔╝███████║
╚═╝     ╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚════╝  ╚═════╝ ╚══════╝`

	b.WriteString(lipgloss.NewStyle().Foreground(accent).Render(banner))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Full Stack Developer · Distributed Systems · Low-level Architecture"))
	b.WriteString("\n\n")

	for i, item := range menuItems {
		prefix := "  "
		if i == len(menuItems)-1 {
			prefix = "  "
		}
		if m.cursor == i {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("%s▶  %s", prefix, item)))
		} else {
			b.WriteString(normalStyle.Render(fmt.Sprintf("%s   %s", prefix, item)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  ↑/↓ navigate · enter select · q quit"))
	return b.String()
}

func (m model) viewAbout() string {
	var b strings.Builder

	b.WriteString(sectionStyle.Render("About Me"))
	b.WriteString("\n\n")

	about := boxStyle.Render(
		lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Render(
			"Hi, I'm Pankaj Kumar — a Full Stack Developer\n" +
				"focused on distributed systems and low-level architecture.\n\n" +
				"I build performant, production-grade systems using\n" +
				"TypeScript, React, Next.js, Rust, Go, and PostgreSQL.\n\n" +
				"I care about what happens at the hardware boundary,\n" +
				"fault-tolerant systems, and clean architecture.",
		),
	)
	b.WriteString(about)
	b.WriteString("\n\n")

	b.WriteString(sectionStyle.Render("Direction"))
	b.WriteString("\n\n")

	directions := []struct{ icon, title, desc string }{
		{"⬡", "Low-level Systems", "Rust and C. Memory, ownership, hardware boundary."},
		{"◈", "Distributed Systems", "Consensus algorithms, state machines, fault tolerance."},
		{"◻", "Clean Architecture", "Simplicity as a feature. Observability. Clear boundaries."},
	}

	for _, d := range directions {
		b.WriteString(lipgloss.NewStyle().Foreground(accent).Bold(true).Render("  "+d.icon+" "+d.title))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("    " + d.desc))
		b.WriteString("\n\n")
	}

	b.WriteString(dimStyle.Render("  esc / q → back to menu"))
	return b.String()
}

func (m model) viewExperience() string {
	var b strings.Builder

	b.WriteString(sectionStyle.Render("Experience"))
	b.WriteString("\n\n")

	for _, job := range jobs {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F3F0FF")).Render("  " + job.Title))
		b.WriteString("  ")
		b.WriteString(badgeStyle.Render("@ " + job.Company))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  " + job.Period))
		b.WriteString("\n\n")

		for _, bullet := range job.Bullets {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB")).Render("    • " + bullet))
			b.WriteString("\n")
		}
		b.WriteString("\n  ")

		for _, t := range job.Tech {
			b.WriteString(tagStyle.Render(t) + " ")
		}
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 58)))
		b.WriteString("\n\n")
	}

	b.WriteString(dimStyle.Render("  esc / q → back to menu"))
	return b.String()
}

func (m model) viewProjects() string {
	var b strings.Builder

	b.WriteString(sectionStyle.Render(fmt.Sprintf("Projects  %s", dimStyle.Render(fmt.Sprintf("[%d/%d]", m.projectIdx+1, len(projects))))))
	b.WriteString("\n\n")

	p := projects[m.projectIdx]

	statusColor := green
	if p.Status == "Building" {
		statusColor = amber
	}

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F3F0FF"))
	b.WriteString(nameStyle.Render("  "+p.Name) + "  ")
	b.WriteString(lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(p.Status))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB")).Width(62).Render("  "+p.Desc))
	b.WriteString("\n\n")

	b.WriteString("  ")
	for _, t := range p.Tech {
		b.WriteString(tagStyle.Render(t) + " ")
	}
	b.WriteString("\n\n")

	if p.URL != "" {
		b.WriteString("  " + linkStyle.Render("→ "+p.URL))
		b.WriteString("\n\n")
	}

	// Mini nav dots
	b.WriteString("  ")
	for i := range projects {
		if i == m.projectIdx {
			b.WriteString(lipgloss.NewStyle().Foreground(accent).Render("● "))
		} else {
			b.WriteString(dimStyle.Render("○ "))
		}
	}
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  ↑/↓ browse projects · esc / q → back"))
	return b.String()
}

func (m model) viewSkills() string {
	var b strings.Builder

	b.WriteString(sectionStyle.Render("Skills"))
	b.WriteString("\n\n")

	order := []string{"Languages", "Frontend", "Backend", "Infra"}
	for _, cat := range order {
		b.WriteString(lipgloss.NewStyle().Foreground(accent).Bold(true).Render("  " + cat))
		b.WriteString("\n  ")
		for _, s := range skills[cat] {
			b.WriteString(tagStyle.Render(s) + " ")
		}
		b.WriteString("\n\n")
	}

	b.WriteString(dimStyle.Render("  esc / q → back to menu"))
	return b.String()
}

func (m model) viewContact() string {
	var b strings.Builder

	b.WriteString(sectionStyle.Render("Contact"))
	b.WriteString("\n\n")

	card := boxStyle.Render(
		lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Render(
			"Hey, you found the terminal version — let's talk.\n\n"+
				"I'm open to new projects, ideas, and opportunities.\n",
		) +
			"\n" +
			lipgloss.NewStyle().Foreground(accent).Render("  Email   ") +
			linkStyle.Render("pankajjaat2608@gmail.com") +
			"\n" +
			lipgloss.NewStyle().Foreground(accent).Render("  GitHub  ") +
			linkStyle.Render("github.com/pankajkumar2608") +
			"\n" +
			lipgloss.NewStyle().Foreground(accent).Render("  LinkedIn") +
			linkStyle.Render("linkedin.com/in/pankaj-jaat/") +
			"\n" +
			lipgloss.NewStyle().Foreground(accent).Render("  Web     ") +
			linkStyle.Render("itzpankaj.site") +
			"\n" +
			lipgloss.NewStyle().Foreground(accent).Render("  X(twitter)     ") +
			linkStyle.Render("https://x.com/itzPankajkoder"),
	)

	b.WriteString(card)
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  esc / q → back to menu"))
	return b.String()
}

// ── SSH Server ────────────────────────────────────────────────────────────────

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort("0.0.0.0", "2323")),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			wishbubble.Middleware(teaHandler),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("PankajOS SSH portfolio running on :2323")
	log.Println("Connect with: ssh localhost -p 2323")
	log.Fatal(s.ListenAndServe())
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	m := initialModel()
	m.width = pty.Window.Width
	m.height = pty.Window.Height
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}