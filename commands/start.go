package commands

import "github.com/xnyo/ugr/common"

// Start handles the /start command
func Start(c *common.Ctx) {
	c.Reply("Hello world")
}
