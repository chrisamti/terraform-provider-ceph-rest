package ceph

import "path"

// PathJoin joins array of interfaces having string and pointer to strings.
func PathJoin(v ...interface{}) string {
	var segments []string
	for _, v := range v {
		switch v := v.(type) {
		case string:
			if v != "" {
				segments = append(segments, v)
			}

		case *string:
			if v != nil {
				if *v != "" {
					segments = append(segments, *v)
				}
			}
		}
	}

	return path.Join(segments...)
}

func StaticCounter() (f func() uint) {
	var i uint
	f = func() uint {
		i++
		return i
	}
	return
}
