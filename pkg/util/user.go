package util

import (
	"fmt"
	"os"
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

func GetCurrentUser() user.User {
	currentUser, err := user.Current()
	if err != nil {
		return user.User{
			Uid:      "0",
			Gid:      "0",
			Username: "root",
			Name:     "root",
			HomeDir:  os.Getenv("HOME"),
		}
	}

	return *currentUser
}
