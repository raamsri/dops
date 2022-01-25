package provider

// SupportedRecordType returns true only for supported record types.
// Currently A, CNAME, SRV, TXT and NS record types are supported.
func SupportedRecordType(recordType string) bool {
	switch recordType {
	case "A", "CNAME", "SRV", "TXT", "NS":
		return true
	default:
		return false
	}
}
