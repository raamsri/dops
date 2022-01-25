package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/toppr-systems/dops/controller"
	"github.com/toppr-systems/dops/dops"
	"github.com/toppr-systems/dops/dops/validation"
	"github.com/toppr-systems/dops/endpoint"
	"github.com/toppr-systems/dops/plan"
	"github.com/toppr-systems/dops/provider"
	"github.com/toppr-systems/dops/provider/aws"
	"github.com/toppr-systems/dops/provider/cloudflare"
	"github.com/toppr-systems/dops/provider/inmemory"
	"github.com/toppr-systems/dops/registry"
	"github.com/toppr-systems/dops/source"
)

func main() {
	cfg := dops.NewConfig()
	if err := cfg.ParseFlags(os.Args[1:]); err != nil {
		log.Fatalf("flag parse error: %v", err)
	}
	log.Infof("config: %s", cfg)

	if err := validation.ValidateConfig(cfg); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	if cfg.LogFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}
	if cfg.DryRun {
		log.Info("dry-run mode, no changes to DNS records will be made")
	}

	ll, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to parse log level: %v", err)
	}
	log.SetLevel(ll)

	ctx, cancel := context.WithCancel(context.Background())

	go serveMetrics(cfg.MetricsAddress)
	go handleSigterm(cancel)

	sourceCfg := &source.Config{
		FQDNTemplate:    cfg.FQDNTemplate,
		ConnectorServer: cfg.ConnectorSourceServer,
		DefaultTargets:  cfg.DefaultTargets,
	}

	// Lookup declared sources to fetch its configuration
	sources, err := source.ByNames(&source.SingletonClientGenerator{}, cfg.Sources, sourceCfg)
	if err != nil {
		log.Fatal(err)
	}

	// Combine multiple sources into a single, deduplicated source
	endpointsSource := source.NewDedupSource(source.NewMultiSource(sources, sourceCfg.DefaultTargets))

	var domainFilter endpoint.DomainFilter
	// RegexDomainFilter overrides DomainFilter
	if cfg.RegexDomainFilter.String() != "" {
		domainFilter = endpoint.NewRegexDomainFilter(cfg.RegexDomainFilter, cfg.RegexDomainExclusion)
	} else {
		domainFilter = endpoint.NewDomainFilterWithExclusions(cfg.DomainFilter, cfg.ExcludeDomains)
	}
	zoneIDFilter := provider.NewZoneIDFilter(cfg.ZoneIDFilter)
	zoneTypeFilter := provider.NewZoneTypeFilter(cfg.AWSZoneType)
	zoneTagFilter := provider.NewZoneTagFilter(cfg.AWSZoneTagFilter)

	var p provider.Provider
	switch cfg.Provider {
	case "aws":
		p, err = aws.NewAWSProvider(
			aws.AWSConfig{
				DomainFilter:         domainFilter,
				ZoneIDFilter:         zoneIDFilter,
				ZoneTypeFilter:       zoneTypeFilter,
				ZoneTagFilter:        zoneTagFilter,
				BatchChangeSize:      cfg.AWSBatchChangeSize,
				BatchChangeInterval:  cfg.AWSBatchChangeInterval,
				EvaluateTargetHealth: cfg.AWSEvaluateTargetHealth,
				AssumeRole:           cfg.AWSAssumeRole,
				APIRetries:           cfg.AWSAPIRetries,
				PreferCNAME:          cfg.AWSPreferCNAME,
				DryRun:               cfg.DryRun,
				ZoneCacheDuration:    cfg.AWSZoneCacheDuration,
			},
		)
	case "cloudflare":
		p, err = cloudflare.NewCloudFlareProvider(domainFilter, zoneIDFilter, cfg.CloudflareZonesPerPage, cfg.CloudflareProxied, cfg.DryRun)
	case "inmemory":
		p, err = inmemory.NewInMemoryProvider(inmemory.InMemoryInitZones(cfg.InMemoryZones), inmemory.InMemoryWithDomain(domainFilter), inmemory.InMemoryWithLogging()), nil
	default:
		log.Fatalf("invalid dns provider: %s", cfg.Provider)
	}
	if err != nil {
		log.Fatal(err)
	}

	var r registry.Registry
	switch cfg.Registry {
	case "noop":
		r, err = registry.NewNoopRegistry(p)
	case "txt":
		r, err = registry.NewTXTRegistry(p, cfg.TXTPrefix, cfg.TXTSuffix, cfg.TXTOwnerID, cfg.TXTCacheInterval, cfg.TXTWildcardReplacement)
	default:
		log.Fatalf("invalid registry: %s", cfg.Registry)
	}
	if err != nil {
		log.Fatal(err)
	}

	policy, exists := plan.Policies[cfg.Policy]
	if !exists {
		log.Fatalf("invalid policy: %s", cfg.Policy)
	}

	ctl := controller.Controller{
		Source:               endpointsSource,
		Registry:             r,
		Policy:               policy,
		Interval:             cfg.Interval,
		DomainFilter:         domainFilter,
		ManagedRecordTypes:   cfg.ManagedDNSRecordTypes,
		MinEventSyncInterval: cfg.MinEventSyncInterval,
	}

	if cfg.Once {
		err := ctl.RunOnce(ctx)
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	if cfg.UpdateEvents {
		ctl.Source.AddEventHandler(ctx, func() { ctl.ScheduleRunOnce(time.Now()) })
	}

	ctl.ScheduleRunOnce(time.Now())
	ctl.Run(ctx)
}

func handleSigterm(cancel func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	<-signals
	log.Info("Received SIGTERM. Terminating...")
	cancel()
}

func serveMetrics(address string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(address, nil))
}
