package wheel

import "maps"

func MergeMaps[K comparable, V any](input ...map[K]V) map[K]V {
	size := 0
	for _, m := range input {
		size += len(m)
	}

	merged := make(map[K]V, size)
	for _, m := range input {
		maps.Copy(merged, m)
	}
	return merged
}
