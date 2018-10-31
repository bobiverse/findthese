package main


// Find in slice
func inSlice(a string, list []string) bool {
	for _, b := range list {
		//fmt.Printf("[%s] == [%s]\n", a, b)
		if b == a {
			return true
		}
	}
	return false
}
