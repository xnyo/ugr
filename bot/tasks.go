package bot

import (
	"log"
	"time"

	"github.com/xnyo/ugr/common"

	"github.com/xnyo/ugr/models"
)

// cleanExpired cleans all expired data:
// - Old orders that still need data
// - Old invites
func cleanExpired() {
	for c := time.Tick(15 * time.Minute); ; {
		log.Printf("Cleaning expired data")
		func() {
			// So that panic() will cause error logging & reporting
			defer common.Recover(recover(), hasSentry)

			// Soft delete orders still pending data after 1h
			if err := Db.Where(
				"status = ? AND updated_at < ?",
				models.OrderStatusNeedsData,
				time.Now().Add(-1*time.Hour),
			).Delete(models.Order{}).Error; err != nil {
				panic(err)
			}

			// Delete invites older than 1h
			if err := Db.Unscoped().Where(
				"created_at < ?",
				time.Now().Add(-1*time.Hour),
			).Delete(models.Invite{}).Error; err != nil {
				panic(err)
			}
		}()
		<-c
	}
}
