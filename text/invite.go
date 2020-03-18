package text

import "fmt"

const invitePrefix = "👋 **Ciao**, sei stato invitato come"
const volunteer = "volontario"
const admin = "amministratore"
const inviteSuffix = "\n\n👇 Fai click qui per accettare! 👇"
const InviteAccept = "✅ Accetta"
const InvitePromptVolunteer = "🛵 Invita come volontario"
const InviteDescriptionVolunteer = "L'utente verrà invitato come volontario"
const InvitePromptAdmin = "🔧 Invita come amministratore"
const InviteDescriptionAdmin = "L'utente verrà invitato come amministratore"

func InviteVolunteer() string {
	return fmt.Sprintf("%s __%s__%s", invitePrefix, volunteer, inviteSuffix)
}

func InviteAdmin() string {
	return fmt.Sprintf("%s __%s__%s", invitePrefix, admin, inviteSuffix)
}
