package executable

func DiscoverExecutables() ([]Executable, error) {
	var result []Executable

	// nix candidates
	nixStoreCandidates := DiscoverNixStoreExecutables(nil)
	result = append(result, nixStoreCandidates...)
	nixShellCandidates := DiscoverNixShellExecutables(nil)
	result = append(result, nixShellCandidates...)

	// exec candidates
	execCandidates := DiscoverPathCandidates(nil)
	result = append(result, execCandidates...)

	// container candidates
	containerCandidates := DiscoverContainerCandidates(nil)
	result = append(result, containerCandidates...)

	return result, nil
}
