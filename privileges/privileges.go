package privileges

// Privileges represents the privileges that a given telegram user has on the bot
type Privileges int

const (
	Normal Privileges = 1 << iota
	Admin
	AdminAddArea
	AdminAddOrder
)
