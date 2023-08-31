package utils

func FilterEmptyStrings(s []string) []string {
	var filtered []string
	for i := range s {
		if s[i] == "" {
			continue
		}
		filtered = append(filtered, s[i])
	}
	return filtered
}

func TrimStringToLen(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
