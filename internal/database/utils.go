package database

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	var result string
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
