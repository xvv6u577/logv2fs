package helper

func CountNodesInUse(nodeStatus map[string]bool) int {
	trueCount := 0
	for _, status := range nodeStatus {
		if status {
			trueCount++
		}
	}
	return trueCount
}
