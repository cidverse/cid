package files

import (
	"encoding/xml"
	"fmt"
)

func ReadXML[T any](content []byte) (T, error) {
	var result T

	err := xml.Unmarshal(content, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal XML content: %w", err)
	}

	return result, nil
}

func WriteXML(content any) ([]byte, error) {
	contentBytes, err := xml.MarshalIndent(content, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML: %w", err)
	}

	return contentBytes, nil
}
