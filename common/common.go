package common

import "regexp"

const defaultHostnameRegex = `esxhostname="([^"]+)"`

func FindHostRegex(s string) string {
	re := regexp.MustCompile(defaultHostnameRegex)
	host := re.FindStringSubmatch(s)
	if len(host) != 2 {
		return ""
	}
	return host[1]
}

func Average(numbers []float64) float64 {
	total := 0.0
	for _, num := range numbers {
		total += num
	}
	return total / float64(len(numbers))
}

func CountOverThreshold(numbers []float64, threshold float64) int {
	count := 0
	for _, num := range numbers {
		if num > threshold {
			count++
		}
	}
	return count
}
