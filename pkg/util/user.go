package util

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

// GetContainerUser returns the user id and group id of the current user to mirror it in the container
func GetContainerUser() string {
	// lookup
	uid := os.Getuid()
	gid := os.Getgid()

	if uid == -1 || gid == -1 {
		if u, err := user.Current(); err == nil {
			if parsedUID, err := strconv.Atoi(u.Uid); err == nil {
				uid = parsedUID
			}
			if parsedGID, err := strconv.Atoi(u.Gid); err == nil {
				gid = parsedGID
			}
		}
	}

	if uid < 0 {
		uid = 1001
	}
	if gid < 0 {
		gid = 0
	}

	// TD-002: GitHub Actions quirk – the current UID/GID can write to the project directory on the host, but file creation fails inside the container, even with the same UID/GID. - UID: 1001, GID: 118
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		uid = 0
		gid = 0
	}

	return fmt.Sprintf("%d:%d", uid, gid)
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
