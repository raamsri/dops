package cloudflare

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"

	"github.com/toppr-systems/dops/endpoint"
	"github.com/toppr-systems/dops/plan"
	"github.com/toppr-systems/dops/provider"
	"github.com/toppr-systems/dops/source"
)

const (
	// cloudFlareCreate is a ChangeAction enum value
	cloudFlareCreate = "CREATE"
	cloudFlareDelete = "DELETE"
	cloudFlareUpdate = "UPDATE"
	// automatic
	defaultCloudFlareRecordTTL = 1
)

var cloudFlareTypeNotSupported = map[string]bool{
	"LOC": true,
	"MX":  true,
	"NS":  true,
	"SPF": true,
	"TXT": true,
	"SRV": true,
}

// cloudFlareDNS is the subset of the CloudFlare API. Add methods as required. Signatures must match exactly.
type cloudFlareDNS interface {
	UserDetails() (cf.User, error)
	ZoneIDByName(zoneName string) (string, error)
	ListZones(zoneID ...string) ([]cf.Zone, error)
	ListZonesContext(ctx context.Context, opts ...cf.ReqOption) (cf.ZonesResponse, error)
	ZoneDetails(zoneID string) (cf.Zone, error)
	DNSRecords(zoneID string, rr cf.DNSRecord) ([]cf.DNSRecord, error)
	CreateDNSRecord(zoneID string, rr cf.DNSRecord) (*cf.DNSRecordResponse, error)
	DeleteDNSRecord(zoneID, recordID string) error
	UpdateDNSRecord(zoneID, recordID string, rr cf.DNSRecord) error
}

type zoneService struct {
	service *cf.API
}

func (z zoneService) UserDetails() (cf.User, error) {
	return z.service.UserDetails()
}

func (z zoneService) ListZones(zoneID ...string) ([]cf.Zone, error) {
	return z.service.ListZones(zoneID...)
}

func (z zoneService) ZoneIDByName(zoneName string) (string, error) {
	return z.service.ZoneIDByName(zoneName)
}

func (z zoneService) CreateDNSRecord(zoneID string, rr cf.DNSRecord) (*cf.DNSRecordResponse, error) {
	return z.service.CreateDNSRecord(zoneID, rr)
}

func (z zoneService) DNSRecords(zoneID string, rr cf.DNSRecord) ([]cf.DNSRecord, error) {
	return z.service.DNSRecords(zoneID, rr)
}
func (z zoneService) UpdateDNSRecord(zoneID, recordID string, rr cf.DNSRecord) error {
	return z.service.UpdateDNSRecord(zoneID, recordID, rr)
}
func (z zoneService) DeleteDNSRecord(zoneID, recordID string) error {
	return z.service.DeleteDNSRecord(zoneID, recordID)
}

func (z zoneService) ListZonesContext(ctx context.Context, opts ...cf.ReqOption) (cf.ZonesResponse, error) {
	return z.service.ListZonesContext(ctx, opts...)
}

func (z zoneService) ZoneDetails(zoneID string) (cf.Zone, error) {
	return z.service.ZoneDetails(zoneID)
}

// CloudFlareProvider is an implementation of Provider for CloudFlare DNS.
type CloudFlareProvider struct {
	provider.BaseProvider
	Client cloudFlareDNS
	// only consider hosted zones managing domains ending in this suffix
	domainFilter      endpoint.DomainFilter
	zoneIDFilter      provider.ZoneIDFilter
	proxiedByDefault  bool
	DryRun            bool
	PaginationOptions cf.PaginationOptions
}

// cloudFlareChange differentiates between ChangeActions
type cloudFlareChange struct {
	Action         string
	ResourceRecord cf.DNSRecord
}

func NewCloudFlareProvider(domainFilter endpoint.DomainFilter, zoneIDFilter provider.ZoneIDFilter, zonesPerPage int, proxiedByDefault bool, dryRun bool) (*CloudFlareProvider, error) {
	// initialize via chosen auth method and return new API object
	var (
		config *cf.API
		err    error
	)
	if os.Getenv("CF_API_TOKEN") != "" {
		config, err = cf.NewWithAPIToken(os.Getenv("CF_API_TOKEN"))
	} else {
		config, err = cf.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudflare provider: %v", err)
	}
	provider := &CloudFlareProvider{
		Client:           zoneService{config},
		domainFilter:     domainFilter,
		zoneIDFilter:     zoneIDFilter,
		proxiedByDefault: proxiedByDefault,
		DryRun:           dryRun,
		PaginationOptions: cf.PaginationOptions{
			PerPage: zonesPerPage,
			Page:    1,
		},
	}
	return provider, nil
}

func (p *CloudFlareProvider) Zones(ctx context.Context) ([]cf.Zone, error) {
	result := []cf.Zone{}
	p.PaginationOptions.Page = 1

	// if there is a zoneIDfilter configured
	// && if the filter isn't just a blank string (used in tests)
	if len(p.zoneIDFilter.ZoneIDs) > 0 && p.zoneIDFilter.ZoneIDs[0] != "" {
		log.Debugln("zoneIDFilter configured, only looking up defined zone IDs")
		for _, zoneID := range p.zoneIDFilter.ZoneIDs {
			log.Debugf("looking up zone %s", zoneID)
			detailResponse, err := p.Client.ZoneDetails(zoneID)
			if err != nil {
				log.Errorf("zone %s lookup failed, %v", zoneID, err)
				continue
			}
			log.WithFields(log.Fields{
				"zoneName": detailResponse.Name,
				"zoneID":   detailResponse.ID,
			}).Debugln("adding zone for consideration")
			result = append(result, detailResponse)
		}
		return result, nil
	}

	log.Debugln("no zoneIDFilter configured, looking at all zones")
	for {
		zonesResponse, err := p.Client.ListZonesContext(ctx, cf.WithPagination(p.PaginationOptions))
		if err != nil {
			return nil, err
		}

		for _, zone := range zonesResponse.Result {
			if !p.domainFilter.Match(zone.Name) {
				log.Debugf("zone %s not in domain filter", zone.Name)
				continue
			}
			result = append(result, zone)
		}
		if p.PaginationOptions.Page == zonesResponse.ResultInfo.TotalPages {
			break
		}
		p.PaginationOptions.Page++
	}
	return result, nil
}

func (p *CloudFlareProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	zones, err := p.Zones(ctx)
	if err != nil {
		return nil, err
	}

	endpoints := []*endpoint.Endpoint{}
	for _, zone := range zones {
		records, err := p.Client.DNSRecords(zone.ID, cf.DNSRecord{})
		if err != nil {
			return nil, err
		}

		// CloudFlare does not support "sets" of targets, but instead returns
		// a single entry for each name/type/target, so group by name
		// and record to allow the planner to calculate the correct plan.
		endpoints = append(endpoints, groupByNameAndType(records)...)
	}

	return endpoints, nil
}

func (p *CloudFlareProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	cloudflareChanges := []*cloudFlareChange{}

	for _, endpoint := range changes.Create {
		for _, target := range endpoint.Targets {
			cloudflareChanges = append(cloudflareChanges, p.newCloudFlareChange(cloudFlareCreate, endpoint, target))
		}
	}

	for i, desired := range changes.UpdateNew {
		current := changes.UpdateOld[i]

		add, remove, leave := provider.Difference(current.Targets, desired.Targets)

		for _, a := range add {
			cloudflareChanges = append(cloudflareChanges, p.newCloudFlareChange(cloudFlareCreate, desired, a))
		}

		for _, a := range leave {
			cloudflareChanges = append(cloudflareChanges, p.newCloudFlareChange(cloudFlareUpdate, desired, a))
		}

		for _, a := range remove {
			cloudflareChanges = append(cloudflareChanges, p.newCloudFlareChange(cloudFlareDelete, current, a))
		}
	}

	for _, endpoint := range changes.Delete {
		for _, target := range endpoint.Targets {
			cloudflareChanges = append(cloudflareChanges, p.newCloudFlareChange(cloudFlareDelete, endpoint, target))
		}
	}

	return p.submitChanges(ctx, cloudflareChanges)
}

func (p *CloudFlareProvider) PropertyValuesEqual(name string, previous string, current string) bool {
	if name == source.CloudflareProxiedKey {
		return plan.CompareBoolean(p.proxiedByDefault, name, previous, current)
	}

	return p.BaseProvider.PropertyValuesEqual(name, previous, current)
}

// submitChanges takes a zone and a collection of Changes and sends them as a single transaction.
func (p *CloudFlareProvider) submitChanges(ctx context.Context, changes []*cloudFlareChange) error {
	// return early if there is nothing to change
	if len(changes) == 0 {
		return nil
	}

	zones, err := p.Zones(ctx)
	if err != nil {
		return err
	}
	// separate into per-zone change sets to be passed to the API.
	changesByZone := p.changesByZone(zones, changes)

	for zoneID, changes := range changesByZone {
		records, err := p.Client.DNSRecords(zoneID, cf.DNSRecord{})
		if err != nil {
			return fmt.Errorf("could not fetch records from zone, %v", err)
		}
		for _, change := range changes {
			logFields := log.Fields{
				"record": change.ResourceRecord.Name,
				"type":   change.ResourceRecord.Type,
				"ttl":    change.ResourceRecord.TTL,
				"action": change.Action,
				"zone":   zoneID,
			}

			log.WithFields(logFields).Info("Changing record.")

			if p.DryRun {
				continue
			}

			if change.Action == cloudFlareUpdate {
				recordID := p.getRecordID(records, change.ResourceRecord)
				if recordID == "" {
					log.WithFields(logFields).Errorf("failed to find previous record: %v", change.ResourceRecord)
					continue
				}
				err := p.Client.UpdateDNSRecord(zoneID, recordID, change.ResourceRecord)
				if err != nil {
					log.WithFields(logFields).Errorf("failed to update record: %v", err)
				}
			} else if change.Action == cloudFlareDelete {
				recordID := p.getRecordID(records, change.ResourceRecord)
				if recordID == "" {
					log.WithFields(logFields).Errorf("failed to find previous record: %v", change.ResourceRecord)
					continue
				}
				err := p.Client.DeleteDNSRecord(zoneID, recordID)
				if err != nil {
					log.WithFields(logFields).Errorf("failed to delete record: %v", err)
				}
			} else if change.Action == cloudFlareCreate {
				_, err := p.Client.CreateDNSRecord(zoneID, change.ResourceRecord)
				if err != nil {
					log.WithFields(logFields).Errorf("failed to create record: %v", err)
				}
			}
		}
	}
	return nil
}

// AdjustEndpoints modifies the endpoints as needed by the specific provider
func (p *CloudFlareProvider) AdjustEndpoints(endpoints []*endpoint.Endpoint) []*endpoint.Endpoint {
	adjustedEndpoints := []*endpoint.Endpoint{}
	for _, e := range endpoints {
		if shouldBeProxied(e, p.proxiedByDefault) {
			e.RecordTTL = 0
		}
		adjustedEndpoints = append(adjustedEndpoints, e)
	}
	return adjustedEndpoints
}

// changesByZone separates a multi-zone change into a single change per zone.
func (p *CloudFlareProvider) changesByZone(zones []cf.Zone, changeSet []*cloudFlareChange) map[string][]*cloudFlareChange {
	changes := make(map[string][]*cloudFlareChange)
	zoneNameIDMapper := provider.ZoneIDName{}

	for _, z := range zones {
		zoneNameIDMapper.Add(z.ID, z.Name)
		changes[z.ID] = []*cloudFlareChange{}
	}

	for _, c := range changeSet {
		zoneID, _ := zoneNameIDMapper.FindZone(c.ResourceRecord.Name)
		if zoneID == "" {
			log.Debugf("Skipping record %s because no hosted zone matching record DNS Name was detected", c.ResourceRecord.Name)
			continue
		}
		changes[zoneID] = append(changes[zoneID], c)
	}

	return changes
}

func (p *CloudFlareProvider) getRecordID(records []cf.DNSRecord, record cf.DNSRecord) string {
	for _, zoneRecord := range records {
		if zoneRecord.Name == record.Name && zoneRecord.Type == record.Type && zoneRecord.Content == record.Content {
			return zoneRecord.ID
		}
	}
	return ""
}

func (p *CloudFlareProvider) newCloudFlareChange(action string, endpoint *endpoint.Endpoint, target string) *cloudFlareChange {
	ttl := defaultCloudFlareRecordTTL
	proxied := shouldBeProxied(endpoint, p.proxiedByDefault)

	if endpoint.RecordTTL.IsConfigured() {
		ttl = int(endpoint.RecordTTL)
	}

	if len(endpoint.Targets) > 1 {
		log.Errorf("Updates should have just one target")
	}

	return &cloudFlareChange{
		Action: action,
		ResourceRecord: cf.DNSRecord{
			Name:    endpoint.DNSName,
			TTL:     ttl,
			Proxied: proxied,
			Type:    endpoint.RecordType,
			Content: target,
		},
	}
}

func shouldBeProxied(endpoint *endpoint.Endpoint, proxiedByDefault bool) bool {
	proxied := proxiedByDefault

	for _, v := range endpoint.ProviderSpecific {
		if v.Name == source.CloudflareProxiedKey {
			b, err := strconv.ParseBool(v.Value)
			if err != nil {
				log.Errorf("Failed to parse annotation [%s]: %v", source.CloudflareProxiedKey, err)
			} else {
				proxied = b
			}
			break
		}
	}

	if cloudFlareTypeNotSupported[endpoint.RecordType] || strings.Contains(endpoint.DNSName, "*") {
		proxied = false
	}
	return proxied
}

func groupByNameAndType(records []cf.DNSRecord) []*endpoint.Endpoint {
	endpoints := []*endpoint.Endpoint{}

	// group supported records by name and type
	groups := map[string][]cf.DNSRecord{}

	for _, r := range records {
		if !provider.SupportedRecordType(r.Type) {
			continue
		}

		groupBy := r.Name + r.Type
		if _, ok := groups[groupBy]; !ok {
			groups[groupBy] = []cf.DNSRecord{}
		}

		groups[groupBy] = append(groups[groupBy], r)
	}

	// create single endpoint with all the targets for each name/type
	for _, records := range groups {
		targets := make([]string, len(records))
		for i, record := range records {
			targets[i] = record.Content
		}
		endpoints = append(endpoints,
			endpoint.NewEndpointWithTTL(
				records[0].Name,
				records[0].Type,
				endpoint.TTL(records[0].TTL),
				targets...).
				WithProviderSpecific(source.CloudflareProxiedKey, strconv.FormatBool(records[0].Proxied)))
	}

	return endpoints
}
