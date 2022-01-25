package provider

import "strings"

type ZoneIDName map[string]string

func (z ZoneIDName) Add(zoneID, zoneName string) {
	z[zoneID] = zoneName
}

func (z ZoneIDName) FindZone(hostname string) (suitableZoneID, suitableZoneName string) {
	for zoneID, zoneName := range z {
		if hostname == zoneName || strings.HasSuffix(hostname, "."+zoneName) {
			if suitableZoneName == "" || len(zoneName) > len(suitableZoneName) {
				suitableZoneID = zoneID
				suitableZoneName = zoneName
			}
		}
	}
	return
}
