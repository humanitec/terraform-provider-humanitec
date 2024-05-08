package provider

type Config struct {
	ApiPrefix string `json:"apiPrefix,omitempty"`
	Org       string `json:"org,omitempty"`
	Token     string `json:"token,omitempty"`
}

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

func findInSlicePtr[T any](in *[]T, f func(T) bool) (T, bool) {
	found := false
	var element T

	if in == nil {
		return element, found
	}

	for _, e := range *in {
		if f(e) {
			found = true
			element = e
			break
		}
	}

	return element, found
}
