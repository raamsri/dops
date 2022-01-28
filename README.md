# dops [DNS Operations]

**`dops`** manages DNS operations on a cloud provider to synchronize the needed hostname-target values(records) from multiple sources.

**Note:** `dops` is *NOT* a DNS server

## Providers

A `provider` is a cloud service provider with supported DNS web service like DNS registration and routing. This is where the DNS records are operated. Currently, supported providers are:

1. `aws` - AWS Route53
2. `cloudflare` - Cloudflare
3. `inmemory` - Emulates a provider for testing

More providers can be individually added when required by implementing the `provider.Provider{}` interface methods.

## Sources

A `source` provides the list of DNS records(*endpoints*) that must be created/synchronised with a suported `provider`.

Supported sources:

1. `connector` - This is a TCP connection source backed by *Go* native *gob* decoding. This is a super convenient approach for any other Go program to transmit DNS records via `net.Listen("tcp", ":xxxx")`. After the connection is established, `dops` reconciliation loop will query the *listening* source for endpoints periodically. This design makes `dops` *pluggable* with another internal service.

2. `dummy` - emulates a source by providing randomly generated endpoints for testing.

3. `empty` - An empty source provides no endpoint; a quick cleanup tool for testing.

More sources can be individually added when required by implementing the `source.Source{}` interface methods.

## Ownership

Multiple instances of `dops` with different filters/parameteres can be operated on a specific or a set of hosted zones. To facilitate harmonious co-operation, for each endpoint dops injects a TXT record with ownership labels pertaining to the running instance that operates that set of records.

Unlike `A` type records, `CNAME` records cannot share hostname with any other type(A, TXT, SRV, etc.) To circumvent this limitation, TXT records are maintained with the user provided prefix. **This is transparently handled, users need not fret**.

For example:

A source endpoint with hostname **try** for domain **dops2.toppr.systems** of type `A` record will create two entries in the provider's DNS hosted zone. Note that the hostnames are same for both type of records.

> pvsb.dops2.toppr.systems
>> Actual record -> `try.dops2.toppr.systems 0 IN A  192.0.2.252`
>>
>>Ownership record -> `try.dops2.toppr.systems 0 IN TXT  \"origin=dops,dops/owner=test\"`

However, the same endpoint of record type `CNAME` will have an ownership record with a different hostname with prefix.

> pvsb.dops2.toppr.systems
>> Actual record -> `try.dops2.toppr.systems 0 IN CNAME ip-10-1-0-80.ap-south-1.compute.internal`
>>
>>Ownership record -> `prefix-try.dops2.toppr.systems 0 IN TXT  \"origin=dops,dops/owner=test\"`

It's not recommended to manually modify dops managed records on the cloud portal, it will leave records in an inconsistent state while synchronising.

## CLI

`dops` takes parameters in effectively two forms - command flags and env variables. Both can be mixed. Parameters marked as *required* are mandatory.

More information is available with `dops --help`

```bash
$ dops --source=dummy \
--provider=aws \
--fqdn-template="dops2.toppr.systems" \
--txt-owner-id=test \
--aws-zone-type=private \
--aws-zones-cache-duration=2h \
--domain-filter="dops2.toppr.systems" \
--domain-filter="dops.toppr.systems" \
--log-format=json \
--log-level=info

INFO[0000] config: {DefaultTargets:[] Sources:[dummy] FQDNTemplate:dops2.toppr.systems PublishHostIP:false ConnectorSourceServer:localhost:9876 Provider:aws DomainFilter:[dops2.toppr.systems dops.toppr.systems] ExcludeDomains:[] RegexDomainFilter: RegexDomainExclusion: ZoneIDFilter:[] AWSZoneType:private AWSZoneTagFilter:[] AWSAssumeRole: AWSBatchChangeSize:1000 AWSBatchChangeInterval:1s AWSEvaluateTargetHealth:true AWSAPIRetries:3 AWSPreferCNAME:true AWSZoneCacheDuration:2h0m0s CloudflareProxied:false CloudflareZonesPerPage:50 InMemoryZones:[] Policy:sync Registry:txt TXTOwnerID:test TXTPrefix: TXTSuffix: Interval:1m0s MinEventSyncInterval:5s Once:false DryRun:false UpdateEvents:false LogFormat:json MetricsAddress::7979 LogLevel:info TXTCacheInterval:0s TXTWildcardReplacement: ManagedDNSRecordTypes:[A CNAME]}


{"level":"info","msg":"Applying provider record filter for domains: [dops2.toppr.systems. .dops2.toppr.systems. dops.toppr.systems. .dops.toppr.systems.]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-ccod.dops2.toppr.systems A [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-ccod.dops2.toppr.systems TXT [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-cdzs.dops2.toppr.systems A [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-cdzs.dops2.toppr.systems TXT [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-gusp.dops2.toppr.systems A [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-gusp.dops2.toppr.systems TXT [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-gyac.dops2.toppr.systems A [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-gyac.dops2.toppr.systems TXT [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-rfig.dops2.toppr.systems A [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"Desired change: CREATE dummy-rfig.dops2.toppr.systems TXT [Id: /hostedzone/Z055666939ELBYE3XXXXX]","time":"2022-01-28T16:26:05+05:30"}
{"level":"info","msg":"10 record(s) in zone dops2.toppr.systems. [Id: /hostedzone/Z055666939ELBYE3XXXXX] were successfully updated","time":"2022-01-28T16:26:06+05:30"}
```

## Contribution

When teams start using dops, it's likely that needs of various other source *(perhaps read endpoints from text file? json over http?)* and providers will arise. More information about design and how-to contribute documents will be added when interest to contribute is shown in our [engienering forum](https://discuss.toppr.systems).
