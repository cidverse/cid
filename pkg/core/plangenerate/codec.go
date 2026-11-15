package plangenerate

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// maxDecodedPlanSize limits the decompressed size of an encoded plan.
const maxDecodedPlanSize = 16 << 20

// EncodePlan serializes a plan as base64(gzip(json)), compact enough to embed in CI configuration (e.g. a GitLab CI/CD variable).
func EncodePlan(plan Plan) (string, error) {
	data, err := json.Marshal(plan)
	if err != nil {
		return "", fmt.Errorf("failed to marshal plan: %w", err)
	}

	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip writer: %w", err)
	}
	if _, err = zw.Write(data); err != nil {
		return "", fmt.Errorf("failed to compress plan: %w", err)
	}
	if err = zw.Close(); err != nil {
		return "", fmt.Errorf("failed to compress plan: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// DecodePlan reverses EncodePlan.
func DecodePlan(encoded string) (Plan, error) {
	compressed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return Plan{}, fmt.Errorf("failed to decode plan: %w", err)
	}

	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return Plan{}, fmt.Errorf("failed to decompress plan: %w", err)
	}
	defer zr.Close()

	data, err := io.ReadAll(io.LimitReader(zr, maxDecodedPlanSize+1))
	if err != nil {
		return Plan{}, fmt.Errorf("failed to decompress plan: %w", err)
	}
	if len(data) > maxDecodedPlanSize {
		return Plan{}, fmt.Errorf("decoded plan exceeds %d bytes", maxDecodedPlanSize)
	}

	var plan Plan
	if err = json.Unmarshal(data, &plan); err != nil {
		return Plan{}, fmt.Errorf("failed to unmarshal plan: %w", err)
	}

	return plan, nil
}
