package text

import "fmt"

// General text
const (
	Unauthorized  = "⛔️ Non puoi usare questa funzionalità"
	SessionError  = "⚠️ **Si è verificato un errore nella sessione corrente**. Per favore, ricomincia."
	ErrorOccurred = "Si è verificato un errore."
	MainMenu      = "👈 Menu principale"
)

// W returns a warning-like error message
func W(s string) string {
	return fmt.Sprintf("⚠️ **%s**", s)
}
