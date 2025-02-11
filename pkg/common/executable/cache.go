package executable

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cidverse/cid/pkg/util"
)

// TODO: freeze constraints only?

type cacheData struct {
	Timestamp  time.Time        `json:"timestamp"`  // Timestamp when the cache was created
	Candidates []typedCandidate `json:"candidates"` // Candidates contains the candidate list
}

type typedCandidate struct {
	Type      string          `json:"type"`      // Type required for unmarshalling
	Candidate json.RawMessage `json:"candidate"` // Candidate contains the actual candidate data
}

var executablesLockFile = filepath.Join(util.CIDStateDir(), "executable-lock.json")

// UpdateExecutableCache persists the candidates into a file
func UpdateExecutableCache(candidates []Candidate) error {
	var result []typedCandidate

	for _, c := range candidates {
		data, err := json.Marshal(c)
		if err != nil {
			return err
		}
		result = append(result, typedCandidate{
			Type:      fmt.Sprintf("%T", c),
			Candidate: data,
		})
	}

	data, err := json.Marshal(cacheData{
		Timestamp:  time.Now(),
		Candidates: result,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(executablesLockFile, data, 0644)
}

// ResetExecutableCache clears the candidate cache
func ResetExecutableCache() {
	_ = os.Remove(executablesLockFile)
}

// LoadCachedExecutables loads the candidates from the cache
func LoadCachedExecutables() ([]Candidate, error) {
	if _, statErr := os.Stat(executablesLockFile); statErr == nil {
		data, err := os.ReadFile(executablesLockFile)
		if err != nil {
			return nil, err
		}

		var cached cacheData
		if err = json.Unmarshal(data, &cached); err != nil {
			return nil, err
		}

		var candidates []Candidate
		for _, c := range cached.Candidates {
			var candidate Candidate
			switch c.Type {
			case "executable.ExecCandidate":
				candidate = &ExecCandidate{}
			case "executable.NixStoreCandidate":
				candidate = &NixStoreCandidate{}
			case "executable.NixShellCandidate":
				candidate = &NixShellCandidate{}
			case "executable.ContainerCandidate":
				candidate = &ContainerCandidate{}
			default:
				return nil, fmt.Errorf("unknown executable type: %s", c.Type)
			}

			if err = json.Unmarshal(c.Candidate, candidate); err != nil {
				return nil, err
			}
			candidates = append(candidates, candidate)
		}

		return candidates, nil
	}

	return nil, nil
}

func LoadExecutables() ([]Candidate, error) {
	executableCandidates, err := LoadCachedExecutables()
	if err != nil {
		return nil, err
	}
	if len(executableCandidates) == 0 {
		// discover executables
		executableCandidates, err = DiscoverExecutables()
		if err != nil {
			return nil, err
		}

		// persist cache
		err = UpdateExecutableCache(executableCandidates)
		if err != nil {
			return nil, err
		}
	}

	return executableCandidates, nil
}
