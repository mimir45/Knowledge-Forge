package config

// merge overlays high onto low and returns the result. low is not modified.
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

// mergeValue recurses only where both sides are maps.
func mergeValue(low, high any) any {
	lm, lok := low.(map[string]any)
	hm, hok := high.(map[string]any)
	if lok && hok {
		return merge(lm, hm)
	}
	return high
}
