package candidate

func DiscoverCandidates() ([]Candidate, error) {
	var result []Candidate

	// exec candidates
	execCandidates := DiscoverPathCandidates(nil)
	result = append(result, execCandidates...)

	// container candidates
	containerCandidates := DiscoverContainerCandidates(nil)
	result = append(result, containerCandidates...)

	// nix candidates
	nixCandidates := DiscoverNixStoreCandidates(nil)
	result = append(result, nixCandidates...)

	return result, nil
}
