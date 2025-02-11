package executable

func DiscoverExecutables() ([]Candidate, error) {
	var result []Candidate

	// nix candidates
	nixCandidates := DiscoverNixStoreCandidates(nil)
	result = append(result, nixCandidates...)

	// exec candidates
	execCandidates := DiscoverPathCandidates(nil)
	result = append(result, execCandidates...)

	// container candidates
	containerCandidates := DiscoverContainerCandidates(nil)
	result = append(result, containerCandidates...)

	return result, nil
}
