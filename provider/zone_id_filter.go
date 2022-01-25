package provider

import "strings"

// ZoneIDFilter holds a list of zone ids to filter by
type ZoneIDFilter struct {
	ZoneIDs []string
}

// NewZoneIDFilter returns a new ZoneIDFilter given a list of zone ids
func NewZoneIDFilter(zoneIDs []string) ZoneIDFilter {
	return ZoneIDFilter{zoneIDs}
}

// Match checks whether a zone matches one of the provided zone ids
func (f ZoneIDFilter) Match(zoneID string) bool {
	// An empty filter includes all zones.
	if len(f.ZoneIDs) == 0 {
		return true
	}

	for _, id := range f.ZoneIDs {
		if strings.HasSuffix(zoneID, id) {
			return true
		}
	}

	return false
}
