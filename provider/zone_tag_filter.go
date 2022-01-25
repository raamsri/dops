package provider

import (
	"strings"
)

// ZoneTagFilter holds a list of zone tags to filter by
type ZoneTagFilter struct {
	zoneTags []string
}

// NewZoneTagFilter returns a new ZoneTagFilter given a list of zone tags
func NewZoneTagFilter(tags []string) ZoneTagFilter {
	if len(tags) == 1 && len(tags[0]) == 0 {
		tags = []string{}
	}
	return ZoneTagFilter{zoneTags: tags}
}

// Match checks whether a zone's set of tags matches the provided tag values
func (f ZoneTagFilter) Match(tagsMap map[string]string) bool {
	for _, tagFilter := range f.zoneTags {
		filterParts := strings.SplitN(tagFilter, "=", 2)
		switch len(filterParts) {
		case 1:
			if _, hasTag := tagsMap[filterParts[0]]; !hasTag {
				return false
			}
		case 2:
			if value, hasTag := tagsMap[filterParts[0]]; !hasTag || value != filterParts[1] {
				return false
			}
		}
	}
	return true
}

// IsEmpty returns true if there are no tags for the filter
func (f ZoneTagFilter) IsEmpty() bool {
	return len(f.zoneTags) == 0
}
