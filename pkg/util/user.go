package util

import (
	"fmt"
	"os/user"
)

// GetContainerUser returns the user id and group id of the current user to mirror it in the container
func GetContainerUser() string {
	result := "1001:0"
	if currentUser, err := user.Current(); err == nil {
		result = fmt.Sprintf("%s:0", currentUser.Uid)
	}

	return result
}
