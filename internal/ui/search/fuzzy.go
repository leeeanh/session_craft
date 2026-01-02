package search

import (
	"strings"
)

// Match returns true if query characters appear in text in order, and the indices of matches.
func Match(query, text string) (bool, []int) {
	if query == "" {
		return true, nil
	}

	lowerQuery := strings.ToLower(query)
	lowerText := strings.ToLower(text)

	qIdx := 0
	qLen := len(lowerQuery)
	var indices []int

	for i := 0; i < len(lowerText); i++ {
		if lowerText[i] == lowerQuery[qIdx] {
			indices = append(indices, i)
			qIdx++
			if qIdx == qLen {
				return true, indices
			}
		}
	}
	return false, nil
}
