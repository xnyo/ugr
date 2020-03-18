package privileges

// Privileges represents the privileges that a given telegram user has on the bot
type Privileges int

// Guest special privileges (can't be used as a flag!)
const Guest Privileges = -1

// User privilege constants
const (
	// Can use volunteer features
	Normal Privileges = 1 << iota

	// Can use the admin panel
	Admin

	// Can manage areas
	AdminManageAreas

	// Can manage orders
	AdminManageOrders
)
