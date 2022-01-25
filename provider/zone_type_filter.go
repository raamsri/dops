package provider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

const (
	zoneTypePublic  = "public"
	zoneTypePrivate = "private"
)

type ZoneTypeFilter struct {
	zoneType string
}

func NewZoneTypeFilter(zoneType string) ZoneTypeFilter {
	return ZoneTypeFilter{zoneType: zoneType}
}

// Match checks whether a zone matches the zone type that's filtered
func (f ZoneTypeFilter) Match(rawZoneType interface{}) bool {
	// An empty zone filter includes all hosted zones.
	if f.zoneType == "" {
		return true
	}

	switch zoneType := rawZoneType.(type) {
	case string:
		switch f.zoneType {
		case zoneTypePublic:
			return zoneType == zoneTypePublic
		case zoneTypePrivate:
			return zoneType == zoneTypePrivate
		}
	case *route53.HostedZone:
		// If the zone has no config, it's assumed as public zone since the config's field
		// `PrivateZone` is false by default
		if zoneType.Config == nil {
			return f.zoneType == zoneTypePublic
		}

		switch f.zoneType {
		case zoneTypePublic:
			return !aws.BoolValue(zoneType.Config.PrivateZone)
		case zoneTypePrivate:
			return aws.BoolValue(zoneType.Config.PrivateZone)
		}
	}

	// Return false on any other path, e.g. unknown zone type filter value
	return false
}
