package sonarqubecommon

import (
	"fmt"
)

func PrepareProject(server string, accessToken string, organization string, projectKey string, projectName string, projectDescription string, mainBranch string) error {
	// query branches
	branchList, branchListErr := GetDefaultBranch(server, accessToken, projectKey)
	if branchListErr != nil {
		// no access or project doesn't exist, create it
		createErr := CreateProject(server, accessToken, organization, projectKey, projectName, mainBranch)
		if createErr != nil {
			return fmt.Errorf("failed to create sonarqube project: %s", createErr.Error())
		}

		return nil
	}

	// find current main branch
	currentMainBranch := ""
	for _, branch := range branchList.Branches {
		if branch.IsMain {
			currentMainBranch = branch.Name
			break
		}
	}

	// rename main branch if needed
	if mainBranch != currentMainBranch {
		// delete possible conflicts
		if containsBranch(branchList.Branches, mainBranch) {
			deleteErr := DeleteBranch(server, accessToken, projectKey, mainBranch)
			if deleteErr != nil {
				return fmt.Errorf("failed to delete branch %s: %s", mainBranch, deleteErr.Error())
			}
		}

		// rename main branch
		renameErr := RenameMainBranch(server, accessToken, projectKey, mainBranch)
		if renameErr != nil {
			return fmt.Errorf("failed to rename main branch: %s", renameErr.Error())
		}
	}

	return nil
}

func containsBranch(branches []*Branch, name string) bool {
	for _, branch := range branches {
		if branch == nil {
			continue
		}

		if branch.Name == name {
			return true
		}
	}

	return false
}
