package utils

// Ticksize format should be 0.xxx
func FindPrecisionFromTickSize(tickSize string) *int {
	if string(tickSize[0]) != "0" || string(tickSize[1]) != "." {
		return nil
	}
	precision := 0
	for i := 2; i < len(tickSize); i++ {
		precision++
		if string(tickSize[i]) == "1" {
			break
		}
	}
	return &precision
}
