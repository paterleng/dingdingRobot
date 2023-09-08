package utils

func DeleteSliceElem(a []interface{}, elem interface{}) []interface{} {
	j := 0
	for _, value := range a {
		if value != elem {
			a[j] = value
			j++
		}
	}
	return a[:j]
}

func DiffArray(a []int, b []int) []int {
	var diffArray []int
	temp := map[int]struct{}{}

	for _, val := range b {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}
