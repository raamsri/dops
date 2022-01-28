package dops

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/toppr-systems/dops/endpoint"

	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"
)

const (
	passwordMask = "******"
)

var (
	// Version is current version of the application, generated at build time
	Version = "unknown"
)

// Config is project-wide configuration
type Config struct {
	DefaultTargets          []string
	Sources                 []string
	FQDNTemplate            string
	PublishHostIP           bool
	ConnectorSourceServer   string
	Provider                string
	DomainFilter            []string
	ExcludeDomains          []string
	RegexDomainFilter       *regexp.Regexp
	RegexDomainExclusion    *regexp.Regexp
	ZoneIDFilter            []string
	AWSZoneType             string
	AWSZoneTagFilter        []string
	AWSAssumeRole           string
	AWSBatchChangeSize      int
	AWSBatchChangeInterval  time.Duration
	AWSEvaluateTargetHealth bool
	AWSAPIRetries           int
	AWSPreferCNAME          bool
	AWSZoneCacheDuration    time.Duration
	CloudflareProxied       bool
	CloudflareZonesPerPage  int
	InMemoryZones           []string
	Policy                  string
	Registry                string
	TXTOwnerID              string
	TXTPrefix               string
	TXTSuffix               string
	Interval                time.Duration
	MinEventSyncInterval    time.Duration
	Once                    bool
	DryRun                  bool
	UpdateEvents            bool
	LogFormat               string
	MetricsAddress          string
	LogLevel                string
	TXTCacheInterval        time.Duration
	TXTWildcardReplacement  string
	ManagedDNSRecordTypes   []string
}

var defaultConfig = &Config{
	DefaultTargets:          []string{},
	Sources:                 nil,
	FQDNTemplate:            "",
	PublishHostIP:           false,
	ConnectorSourceServer:   "localhost:9876",
	Provider:                "",
	DomainFilter:            []string{},
	ExcludeDomains:          []string{},
	RegexDomainFilter:       regexp.MustCompile(""),
	RegexDomainExclusion:    regexp.MustCompile(""),
	ZoneIDFilter:            []string{},
	AWSZoneType:             "",
	AWSZoneTagFilter:        []string{},
	AWSAssumeRole:           "",
	AWSBatchChangeSize:      1000,
	AWSBatchChangeInterval:  time.Second,
	AWSEvaluateTargetHealth: true,
	AWSAPIRetries:           3,
	AWSPreferCNAME:          false,
	AWSZoneCacheDuration:    0 * time.Second,
	CloudflareProxied:       false,
	CloudflareZonesPerPage:  50,
	InMemoryZones:           []string{},
	Policy:                  "sync",
	Registry:                "txt",
	TXTOwnerID:              "default",
	TXTPrefix:               "",
	TXTSuffix:               "",
	TXTCacheInterval:        0,
	TXTWildcardReplacement:  "",
	MinEventSyncInterval:    5 * time.Second,
	Interval:                time.Minute,
	Once:                    false,
	DryRun:                  false,
	UpdateEvents:            false,
	LogFormat:               "text",
	MetricsAddress:          ":7979",
	LogLevel:                logrus.InfoLevel.String(),
	ManagedDNSRecordTypes:   []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME},
}

// NewConfig returns new Config object
func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) String() string {
	// prevent logging of sensitive information
	temp := *cfg

	t := reflect.TypeOf(temp)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if val, ok := f.Tag.Lookup("secure"); ok && val == "yes" {
			if f.Type.Kind() != reflect.String {
				continue
			}
			v := reflect.ValueOf(&temp).Elem().Field(i)
			if v.String() != "" {
				v.SetString(passwordMask)
			}
		}
	}

	return fmt.Sprintf("%+v", temp)
}

// allLogLevelsAsStrings returns all logrus levels as a list of strings
func allLogLevelsAsStrings() []string {
	var levels []string
	for _, level := range logrus.AllLevels {
		levels = append(levels, level.String())
	}
	return levels
}

// ParseFlags adds and parses flags from command line using kingpin
func (cfg *Config) ParseFlags(args []string) error {
	boot := kingpin.New("dops", "DNSOps synchronizes DNS records of one or more sources with DNS providers.\n\nAll flags may be replaced with env vars - `--example-flag` -> `DOPS_EXAMPLE_FLAG=1` or `--example-flag value` -> `DOPS_EXAMPLE_FLAG=value`")
	boot.Version(Version)
	boot.DefaultEnvars()

	// Sources
	boot.Flag("source", "The resource types that are queried for endpoints; specify multiple times for multiple sources (required, options: dummy, connector, empty)").Required().PlaceHolder("source").EnumsVar(&cfg.Sources, "dummy", "connector", "empty")
	boot.Flag("fqdn-template", "A templated string that's used to generate DNS names from sources that don't define a hostname themselves, or to add a hostname suffix when paired with the dummy source (optional)").Default(defaultConfig.FQDNTemplate).StringVar(&cfg.FQDNTemplate)
	boot.Flag("managed-record-types", "Comma separated list of record types to manage (default: A, CNAME) (supported records: CNAME, A, NS").Default("A", "CNAME").StringsVar(&cfg.ManagedDNSRecordTypes)
	boot.Flag("default-targets", "Set globally default IP address that will apply as a target instead of source addresses. Specify multiple times for multiple targets (optional)").StringsVar(&cfg.DefaultTargets)
	boot.Flag("connector-source-server", "The server to connect for connector source, valid only when using connector source").Default(defaultConfig.ConnectorSourceServer).StringVar(&cfg.ConnectorSourceServer)
	boot.Flag("publish-host-ip", "Allow dops to publish host-ip for headless services (optional)").BoolVar(&cfg.PublishHostIP)

	// Providers
	boot.Flag("provider", "The DNS provider where the DNS records will be created (required, options: aws, cloudflare, inmemory)").Required().PlaceHolder("provider").EnumVar(&cfg.Provider, "aws", "cloudflare", "inmemory")
	boot.Flag("domain-filter", "Limit possible target zones by a domain suffix; specify multiple times for multiple domains (optional)").Default("").StringsVar(&cfg.DomainFilter)
	boot.Flag("exclude-domains", "Exclude subdomains (optional)").Default("").StringsVar(&cfg.ExcludeDomains)
	boot.Flag("regex-domain-filter", "Limit possible domains and target zones by a Regex filter; Overrides domain-filter (optional)").Default(defaultConfig.RegexDomainFilter.String()).RegexpVar(&cfg.RegexDomainFilter)
	boot.Flag("regex-domain-exclusion", "Regex filter that excludes domains and target zones matched by regex-domain-filter (optional)").Default(defaultConfig.RegexDomainExclusion.String()).RegexpVar(&cfg.RegexDomainExclusion)
	boot.Flag("zone-id-filter", "Filter target zones by hosted zone id; specify multiple times for multiple zones (optional)").Default("").StringsVar(&cfg.ZoneIDFilter)

	boot.Flag("aws-zone-type", "When using the AWS provider, filter for zones of this type (optional, options: public, private)").Default(defaultConfig.AWSZoneType).EnumVar(&cfg.AWSZoneType, "", "public", "private")
	boot.Flag("aws-zone-tags", "When using the AWS provider, filter for zones with these tags").Default("").StringsVar(&cfg.AWSZoneTagFilter)
	boot.Flag("aws-assume-role", "When using the AWS provider, assume this IAM role. Useful for hosted zones in another AWS account. Specify the full ARN, e.g. `arn:aws:iam::123455567:role/dops` (optional)").Default(defaultConfig.AWSAssumeRole).StringVar(&cfg.AWSAssumeRole)
	boot.Flag("aws-batch-change-size", "When using the AWS provider, set the maximum number of changes that will be applied in each batch.").Default(strconv.Itoa(defaultConfig.AWSBatchChangeSize)).IntVar(&cfg.AWSBatchChangeSize)
	boot.Flag("aws-batch-change-interval", "When using the AWS provider, set the interval between batch changes.").Default(defaultConfig.AWSBatchChangeInterval.String()).DurationVar(&cfg.AWSBatchChangeInterval)
	boot.Flag("aws-evaluate-target-health", "When using the AWS provider, set whether to evaluate the health of a DNS target (default: enabled, disable with --no-aws-evaluate-target-health)").Default(strconv.FormatBool(defaultConfig.AWSEvaluateTargetHealth)).BoolVar(&cfg.AWSEvaluateTargetHealth)
	boot.Flag("aws-api-retries", "When using the AWS provider, set the maximum number of retries for API calls before giving up.").Default(strconv.Itoa(defaultConfig.AWSAPIRetries)).IntVar(&cfg.AWSAPIRetries)
	boot.Flag("aws-prefer-cname", "When using the AWS provider, prefer using CNAME instead of ALIAS (default: disabled)").BoolVar(&cfg.AWSPreferCNAME)
	boot.Flag("aws-zones-cache-duration", "When using the AWS provider, set the zones list cache TTL (0s to disable).").Default(defaultConfig.AWSZoneCacheDuration.String()).DurationVar(&cfg.AWSZoneCacheDuration)

	boot.Flag("cloudflare-proxied", "When using the Cloudflare provider, specify if the proxy mode must be enabled (default: disabled)").BoolVar(&cfg.CloudflareProxied)
	boot.Flag("cloudflare-zones-per-page", "When using the Cloudflare provider, specify how many zones per page listed, max. possible 50 (default: 50)").Default(strconv.Itoa(defaultConfig.CloudflareZonesPerPage)).IntVar(&cfg.CloudflareZonesPerPage)

	boot.Flag("inmemory-zone", "Provide a list of pre-configured zones for the inmemory provider; specify multiple times for multiple zones (optional)").Default("").StringsVar(&cfg.InMemoryZones)

	// Policies
	boot.Flag("policy", "Modify how DNS records are synchronized between sources and providers (default: sync, options: sync, upsert-only, create-only)").Default(defaultConfig.Policy).EnumVar(&cfg.Policy, "sync", "upsert-only", "create-only")

	// Registry
	boot.Flag("registry", "The registry implementation to use to keep track of DNS record ownership (default: txt, options: txt, noop)").Default(defaultConfig.Registry).EnumVar(&cfg.Registry, "txt", "noop")
	boot.Flag("txt-owner-id", "When using the TXT registry, a name that identifies this instance of DNSOps (default: default)").Default(defaultConfig.TXTOwnerID).StringVar(&cfg.TXTOwnerID)
	boot.Flag("txt-prefix", "When using the TXT registry, a custom string that's prefixed to each ownership DNS record (optional). Mutually exclusive with txt-suffix.").Default(defaultConfig.TXTPrefix).StringVar(&cfg.TXTPrefix)
	boot.Flag("txt-suffix", "When using the TXT registry, a custom string that's suffixed to the host portion of each ownership DNS record (optional). Mutually exclusive with txt-prefix.").Default(defaultConfig.TXTSuffix).StringVar(&cfg.TXTSuffix)
	boot.Flag("txt-wildcard-replacement", "When using the TXT registry, a custom string that's used instead of an asterisk for TXT records corresponding to wildcard DNS records (optional)").Default(defaultConfig.TXTWildcardReplacement).StringVar(&cfg.TXTWildcardReplacement)

	// Control loop
	boot.Flag("txt-cache-interval", "The interval between cache synchronizations in duration format (default: disabled)").Default(defaultConfig.TXTCacheInterval.String()).DurationVar(&cfg.TXTCacheInterval)
	boot.Flag("interval", "The interval between two consecutive synchronizations in duration format (default: 1m)").Default(defaultConfig.Interval.String()).DurationVar(&cfg.Interval)
	boot.Flag("min-event-sync-interval", "The minimum interval between two consecutive synchronizations triggered from watch events in duration format (default: 5s)").Default(defaultConfig.MinEventSyncInterval.String()).DurationVar(&cfg.MinEventSyncInterval)
	boot.Flag("once", "When enabled, exits the synchronization loop after the first iteration (default: disabled)").BoolVar(&cfg.Once)
	boot.Flag("dry-run", "When enabled, prints DNS record changes rather than actually performing them (default: disabled)").BoolVar(&cfg.DryRun)
	boot.Flag("events", "When enabled, in addition to running every interval, the reconciliation loop will get triggered when supported sources change (default: disabled)").BoolVar(&cfg.UpdateEvents)

	// Misc.
	boot.Flag("metrics-address", "Address to serve metrics and health check (default: :7979)").Default(defaultConfig.MetricsAddress).StringVar(&cfg.MetricsAddress)
	boot.Flag("log-format", "The format in which log messages are printed (default: text, options: text, json)").Default(defaultConfig.LogFormat).EnumVar(&cfg.LogFormat, "text", "json")
	boot.Flag("log-level", "Set the level of logging. (default: info, options: panic, debug, info, warning, error, fatal").Default(defaultConfig.LogLevel).EnumVar(&cfg.LogLevel, allLogLevelsAsStrings()...)

	_, err := boot.Parse(args)
	if err != nil {
		return err
	}

	return nil
}
