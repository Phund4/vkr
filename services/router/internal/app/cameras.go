package app

import (
	"sort"

	"router/internal/core/domain"
)

// camerasEqual сравнивает списки камер (порядок не важен).
func camerasEqual(a, b []domain.Camera) bool {
	if len(a) != len(b) {
		return false
	}
	ka := cameraKeys(a)
	kb := cameraKeys(b)
	for i := range ka {
		if ka[i] != kb[i] {
			return false
		}
	}
	return true
}

func cameraKeys(cams []domain.Camera) []string {
	out := make([]string, 0, len(cams))
	for _, c := range cams {
		out = append(out, c.SegmentID+"|"+c.CameraID+"|"+c.RTSPURL)
	}
	sort.Strings(out)
	return out
}
