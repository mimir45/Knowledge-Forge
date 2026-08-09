package config

// merge overlays high onto low and returns the result. low is not modified.
//
// The rule is: **maps merge key by key, scalars and lists replace wholesale.** It is the
// only rule a user can predict without reading this file. Union semantics on lists would
// mean a project could never *narrow* one — `static.languages: [go]` in a Go repo would
// silently inherit java, kotlin, python and typescript from the packaged layer, and
// `research.deny_domains` could never be shortened. Narrowing is the common case.
//
// The map case is what makes `pipeline` work: the packaged layer says
// `recall: {engine: none}` and a user layer that only sets `verify:` must not delete it,
// while a user layer setting `verify: {fallback: host}` must keep §E's `engine: advisor`.
func merge(low, high map[string]any) map[string]any {
	out := make(map[string]any, len(low)+len(high))
	for k, v := range low {
		out[k] = v
	}
	for k, hv := range high {
		lv, ok := out[k]
		if !ok {
			out[k] = hv
			continue
		}
		out[k] = mergeValue(lv, hv)
	}
	return out
}

// mergeValue recurses only where both sides are maps. A map replaced by a scalar, or a
// scalar replaced by a map, takes the higher layer whole — a type change is a rewrite,
// not a merge, and guessing otherwise would produce a config neither layer describes.
func mergeValue(low, high any) any {
	lm, lok := low.(map[string]any)
	hm, hok := high.(map[string]any)
	if lok && hok {
		return merge(lm, hm)
	}
	return high
}
