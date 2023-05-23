package util

// GetContainerUser returns the user id and group id of the current user to mirror it in the container
// TODO: This is not working yet because of some issues, we still require root as of now
func GetContainerUser() string {
	return "0:0"
	/*
		if runtime.GOOS == "windows" {
			return "1001:0"
		}

		result := "1001:0"
		if currentUser, err := user.Current(); err == nil {
			result = fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid)
		}

		return result
	*/
}
