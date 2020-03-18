package text

import "fmt"

// Invite prompts constants
const (
	InvitePromptVolunteer      = "🛵 Invita come volontario"
	InviteDescriptionVolunteer = "L'utente verrà invitato come volontario"
	InvitePromptAdmin          = "🔧 Invita come amministratore"
	InviteDescriptionAdmin     = "L'utente verrà invitato come amministratore"
	invitePrefix               = "👋 **Ciao**, sei stato invitato come"
	volunteer                  = "volontario"
	admin                      = "amministratore"
	inviteSuffix               = "\n\n👇 Fai click qui per accettare! 👇"
	InviteAccept               = "✅ Accetta"
)

func invite(what string) string { return fmt.Sprintf("%s __%s__%s", invitePrefix, what, inviteSuffix) }
func InviteVolunteer() string   { return invite(volunteer) }
func InviteAdmin() string       { return invite(admin) }
