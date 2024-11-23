package util

import (
	"strings"

	"github.com/google/uuid"
)

func RandomUUIDWithoutDashes() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
