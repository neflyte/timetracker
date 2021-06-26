package utils

// TrimWithEllipsis trims a string if it is longer than trimLength and appends an ellipsis if the string was trimmed
func TrimWithEllipsis(toTrim string, trimLength int) string {
	if len(toTrim) <= trimLength {
		return toTrim
	}
	return toTrim[0:trimLength-2] + `…`
}
