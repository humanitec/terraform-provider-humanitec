package provider

func valueAtPath[T any](input map[string]interface{}, path []string) (T, bool) {
	lenPath := len(path)

	var v T
	var ok bool

	for i := 0; i < lenPath; i++ {
		elem := input[path[i]]
		if i+1 == lenPath {
			v, ok = elem.(T)
			return v, ok
		}

		input, ok = elem.(map[string]interface{})
		if !ok {
			return v, ok
		}
	}

	return v, ok
}
