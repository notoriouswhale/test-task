package utils

func CalculateTotalPages(total int, pageSize int) int {
	if total == 0 || pageSize == 0 {
		return 0
	}
	return (total + pageSize - 1) / (pageSize)
}
