package plan

// Policy allows to apply different rules to a set of changes.
type Policy interface {
	Apply(changes *Changes) *Changes
}

// Policies is a registry of available policies.
var Policies = map[string]Policy{
	"sync":        &SyncPolicy{},
	"upsert-only": &UpsertOnlyPolicy{},
	"create-only": &CreateOnlyPolicy{},
}

// SyncPolicy allows for full synchronization of DNS records.
type SyncPolicy struct{}

// Apply applies the sync policy which returns the set of changes as is.
func (p *SyncPolicy) Apply(changes *Changes) *Changes {
	return changes
}

// UpsertOnlyPolicy allows everything but deleting DNS records.
type UpsertOnlyPolicy struct{}

// Apply applies the upsert-only policy which strips out any deletions.
func (p *UpsertOnlyPolicy) Apply(changes *Changes) *Changes {
	return &Changes{
		Create:    changes.Create,
		UpdateOld: changes.UpdateOld,
		UpdateNew: changes.UpdateNew,
	}
}

// CreateOnlyPolicy allows only creating DNS records.
type CreateOnlyPolicy struct{}

// Apply applies the create-only policy which strips out updates and deletions.
func (p *CreateOnlyPolicy) Apply(changes *Changes) *Changes {
	return &Changes{
		Create: changes.Create,
	}
}
