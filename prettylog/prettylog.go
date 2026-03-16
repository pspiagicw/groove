package prettylog

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	outWriter io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr

	accentColor  = lipgloss.AdaptiveColor{Light: "25", Dark: "39"}
	mutedColor   = lipgloss.AdaptiveColor{Light: "241", Dark: "245"}
	borderColor  = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}
	successColor = lipgloss.AdaptiveColor{Light: "28", Dark: "42"}
	warnColor    = lipgloss.AdaptiveColor{Light: "172", Dark: "214"}
	errorColor   = lipgloss.AdaptiveColor{Light: "160", Dark: "203"}

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1)

	messageStyle = lipgloss.NewStyle()

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	keyStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Width(18)

	valueStyle = lipgloss.NewStyle().Bold(true)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)
)

func line(label string, labelColor lipgloss.TerminalColor, format string, args ...any) string {
	badge := labelStyle.Background(labelColor).Render(strings.ToUpper(label))
	msg := messageStyle.Render(fmt.Sprintf(format, args...))

	return lipgloss.JoinHorizontal(lipgloss.Top, badge, " ", msg)
}

func writeLine(w io.Writer, rendered string) {
	fmt.Fprintln(w, rendered)
}

func Infof(format string, args ...any) {
	writeLine(outWriter, line("info", accentColor, format, args...))
}

func Successf(format string, args ...any) {
	writeLine(outWriter, line("ok", successColor, format, args...))
}

func Warnf(format string, args ...any) {
	writeLine(errWriter, line("warn", warnColor, format, args...))
}

func Errorf(format string, args ...any) {
	writeLine(errWriter, line("error", errorColor, format, args...))
}

func Fatalf(format string, args ...any) {
	writeLine(errWriter, line("fatal", errorColor, format, args...))
	os.Exit(1)
}

func Title(title string) string {
	return titleStyle.Render(title)
}

func Section(title string) string {
	return sectionStyle.Render(title)
}

func KV(key string, value any) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		keyStyle.Render(key),
		valueStyle.Render(fmt.Sprint(value)),
	)
}

func BoolKV(key string, value bool) string {
	status := "disabled"
	statusStyle := valueStyle.Foreground(errorColor)
	if value {
		status = "enabled"
		statusStyle = valueStyle.Foreground(successColor)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		keyStyle.Render(key),
		statusStyle.Render(status),
	)
}

func Block(title string, rows ...string) string {
	content := []string{Title(title)}
	if len(rows) > 0 {
		content = append(content, "")
		content = append(content, rows...)
	}

	return cardStyle.Render(strings.Join(content, "\n"))
}

func PrintBlock(w io.Writer, title string, rows ...string) {
	fmt.Fprintln(w, Block(title, rows...))
}

func Prompt(label string) string {
	return promptStyle.Render(label)
}
