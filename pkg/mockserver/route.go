package mockserver

type (
	route struct {
		path     string
		segments []routeSegment
		static   bool
		ops      map[string]*operation
	}

	routeSegment struct {
		value string
		param bool
	}
)

func (rt *route) match(path string) (map[string]string, bool) {
	parts := splitPath(path)
	if len(parts) != len(rt.segments) {
		return nil, false
	}

	params := make(map[string]string)

	for idx, segment := range rt.segments {
		part := parts[idx]
		if segment.param {
			params[segment.value] = part
			continue
		}

		if segment.value != part {
			return nil, false
		}
	}

	return params, true
}
