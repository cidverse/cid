package appcommon

import (
	"fmt"
	"hash/fnv"
)

// GenerateCron returns a cron expression based on a schedule ("daily", "weekly", ...) and a seed for deterministic spread.
func GenerateCron(schedule, seed string) string {
	const minMinute = 0
	const maxMinute = 119
	const baseHour = 0

	h := fnv.New32a()
	h.Write([]byte(seed))
	seedVal := h.Sum32()

	offset := int(seedVal % (maxMinute - minMinute + 1))
	hour := baseHour + (offset / 60)
	minute := offset % 60

	switch schedule {
	case "daily":
		return fmt.Sprintf("%d %d * * *", minute, hour)
	case "weekly":
		weekday := int(seedVal % 7)
		return fmt.Sprintf("%d %d * * %d", minute, hour, weekday)
	case "monthly":
		day := int(seedVal%28) + 1
		return fmt.Sprintf("%d %d %d * *", minute, hour, day)
	default:
		panic("generateCron: invalid schedule type (use 'daily' or 'weekly')")
	}
}
