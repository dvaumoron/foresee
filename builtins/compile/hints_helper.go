package compile

var knownLibrary map[string]string

func init() {
	counters := map[string]int{}
	for _, name := range standardLibraryHints {
		counters[name]++
	}

	size := len(standardLibraryHints)
	excluded := map[string]struct{}{}
	for name, counter := range counters {
		if counter > 1 {
			excluded[name] = struct{}{}
			size -= counter
		}
	}

	knownLibrary = make(map[string]string, size)
	for path, name := range standardLibraryHints {
		if _, ok := excluded[name]; !ok {
			knownLibrary[name] = path
		}
	}
}
