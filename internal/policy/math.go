package policy

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}
