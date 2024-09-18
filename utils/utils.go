package utils

import (
	"strconv"
)

// ParseNumWorkers преобразует текст в количество потоков
func ParseNumWorkers(text string) int {
	num, err := strconv.Atoi(text)
	if err != nil || num <= 0 {
		return 1
	}
	return num
}
