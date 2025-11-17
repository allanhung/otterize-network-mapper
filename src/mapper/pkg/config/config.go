package config

import (
	"github.com/amit7itz/goset"
	sharedconfig "github.com/otterize/network-mapper/src/shared/config"
	"github.com/otterize/network-mapper/src/shared/kubeutils"
	"github.com/spf13/viper"
	"time"
)

const (
	ClusterDomainKey                         = "cluster-domain"
	ClusterDomainDefault                     = kubeutils.DefaultClusterDomain
	CloudApiAddrKey                          = "api-address"
	CloudApiAddrDefault                      = "https://app.otterize.com/api"
	UploadIntervalSecondsKey                 = "upload-interval-seconds"
	UploadIntervalSecondsDefault             = 60
	UploadBatchSizeKey                       = "upload-batch-size"
	UploadBatchSizeDefault                   = 500
	ExcludedNamespacesKey                    = "exclude-namespaces"
	OTelEnabledKey                           = "enable-otel-export"
	OTelEnabledDefault                       = false
	OTelMetricKey                            = "otel-metric-name"
	OTelMetricDefault                        = "traces_service_graph_request_total" // same as expected in otel-collector-contrib's servicegraphprocessor
	ExternalTrafficCaptureEnabledKey         = "capture-external-traffic-enabled"
	ExternalTrafficCaptureEnabledDefault     = true
	CreateWebhookCertificateKey              = "create-webhook-certificate"
	CreateWebhookCertificateDefault          = true
	DNSCacheItemsMaxCapacityKey              = "dns-cache-items-max-capacity"
	DNSCacheItemsMaxCapacityDefault          = 100000
	DNSClientIntentsUpdateIntervalKey        = "dns-client-intents-update-interval"
	DNSClientIntentsUpdateIntervalDefault    = 100 * time.Millisecond
	DNSClientIntentsUpdateEnabledKey         = "dns-client-intents-update-enabled"
	DNSClientIntentsUpdateEnabledDefault     = true
	ServiceCacheTTLDurationKey               = "service-cache-ttl-duration"
	ServiceCacheTTLDurationDefault           = 1 * time.Minute
	ServiceCacheSizeKey                      = "service-cache-size"
	ServiceCacheSizeDefault                  = 10000
	MetricsCollectionTrafficCacheSizeKey     = "metrics-collection-traffic-cache-size"
	MetricsCollectionTrafficCacheSizeDefault = 10000
	WebhookServicesCacheSizeKey              = "webhook-services-cache-size"
	WebhookServicesCacheSizeDefault          = 10

	EnableIstioCollectionKey                  = "enable-istio-collection"
	EnableIstioCollectionDefault              = false
	IstioRestrictCollectionToNamespace        = "istio-restrict-collection-to-namespace"
	IstioReportIntervalKey                    = "istio-report-interval"
	IstioReportIntervalDefault                = 30 * time.Second
	IstioCooldownIntervalKey                  = "istio-cooldown-interval"
	IstioCooldownIntervalDefault              = 15 * time.Second
	MetricFetchTimeoutKey                     = "istio-metric-fetch-timeout"
	MetricFetchTimeoutDefault                 = 10 * time.Second
	TimeServerHasToLiveBeforeWeTrustItKey     = "time-server-has-to-live-before-we-trust-it"
	TimeServerHasToLiveBeforeWeTrustItDefault = 5 * time.Minute

	ControlPlaneIPv4CidrPrefixLength        = "control-plane-ipv4-cidr-prefix-length"
	ControlPlaneIPv4CidrPrefixLengthDefault = 32

	TCPDestResolveOnlyControlPlaneByIp        = "tcp-dest-resolve-only-control-plane-by-ip"
	TCPDestResolveOnlyControlPlaneByIpDefault = true
	HttpIdleTimeoutKey                        = "http-idle-timeout"
	HttpIdleTimeoutDefault                    = 30 * time.Second
	HttpReadTimeoutKey                        = "http-read-timeout"
	HttpReadTimeoutDefault                    = 10 * time.Second
	HttpWriteTimeoutKey                       = "http-write-timeout"
	HttpWriteTimeoutDefault                   = 10 * time.Second
	ClientIgnoreListByNameKey                 = "client-ignore-list-by-name"
	ClientIgnoreListByNameDefault             = "coredns"
	ClusterKey                                = "cluster"
	ClusterDefault                            = "cluster.local"
	DbEnabledKey                              = "db-enabled"
	DbEnabledDefault                          = true
	DbHostKey                                 = "db-host"
	DbHostDefault                             = "127.0.0.1"
	DbUsernameKey                             = "db-username"
	DbUsernameDefault                         = "root"
	DbPasswordKey                             = "db-password"
	DbPasswordDefault                         = "password"
	DbPortKey                                 = "db-port"
	DbPortDefault                             = "3306"
	DbDatabaseKey                             = "db-database"
	DbDatabaseDefault                         = "otterise"
	GhaDispatchEnabledKey                     = "gha-dispatch-enabled"
	GhaDispatchEnabledDefault                 = true
	GhaTokenKey                               = "gha-token"
	GhaTokenDefault                           = ""
	GhaUrlKey                                 = "gha-url"
	GhaUrlDefault                             = "api.github.com"
	GhaOwnerKey                               = "gha-owner"
	GhaOwnerDefault                           = ""
	GhaRepoKey                                = "gha-repo"
	GhaRepoDefault                            = ""
	GhaEventTypeKey                           = "gha-event-type"
	GhaEventTypeDefault                       = "recieveNewIntents"
	ExternalIntentsRetentionDaysKey           = "external-intents-retention-days"
	ExternalIntentsRetentionDaysDefault       = 90
)

var excludedNamespaces *goset.Set[string]

func ExcludedNamespaces() *goset.Set[string] {
	return excludedNamespaces
}

func init() {
	viper.SetDefault(sharedconfig.DebugKey, sharedconfig.DebugDefault)
	viper.SetDefault(ClusterDomainKey, ClusterDomainDefault) // If not set by the user, the main.go of mapper will try to find the cluster domain and set it itself.
	viper.SetDefault(CloudApiAddrKey, CloudApiAddrDefault)
	viper.SetDefault(UploadIntervalSecondsKey, UploadIntervalSecondsDefault)
	viper.SetDefault(UploadBatchSizeKey, UploadBatchSizeDefault)
	viper.SetDefault(OTelEnabledKey, OTelEnabledDefault)
	viper.SetDefault(OTelMetricKey, OTelMetricDefault)
	viper.SetDefault(ExternalTrafficCaptureEnabledKey, ExternalTrafficCaptureEnabledDefault)
	viper.SetDefault(CreateWebhookCertificateKey, CreateWebhookCertificateDefault)
	viper.SetDefault(DNSCacheItemsMaxCapacityKey, DNSCacheItemsMaxCapacityDefault)
	viper.SetDefault(DNSClientIntentsUpdateIntervalKey, DNSClientIntentsUpdateIntervalDefault)
	viper.SetDefault(DNSClientIntentsUpdateEnabledKey, DNSClientIntentsUpdateEnabledDefault)
	viper.SetDefault(IstioReportIntervalKey, IstioReportIntervalDefault)
	viper.SetDefault(MetricFetchTimeoutKey, MetricFetchTimeoutDefault)
	viper.SetDefault(IstioCooldownIntervalKey, IstioCooldownIntervalDefault)
	viper.SetDefault(IstioRestrictCollectionToNamespace, "")
	viper.SetDefault(EnableIstioCollectionKey, EnableIstioCollectionDefault)
	viper.SetDefault(ServiceCacheTTLDurationKey, ServiceCacheTTLDurationDefault)
	viper.SetDefault(ServiceCacheSizeKey, ServiceCacheSizeDefault)
	viper.SetDefault(MetricsCollectionTrafficCacheSizeKey, MetricsCollectionTrafficCacheSizeDefault)
	viper.SetDefault(WebhookServicesCacheSizeKey, WebhookServicesCacheSizeDefault)
	viper.SetDefault(TimeServerHasToLiveBeforeWeTrustItKey, TimeServerHasToLiveBeforeWeTrustItDefault)
	viper.SetDefault(ControlPlaneIPv4CidrPrefixLength, ControlPlaneIPv4CidrPrefixLengthDefault)
	viper.SetDefault(TCPDestResolveOnlyControlPlaneByIp, TCPDestResolveOnlyControlPlaneByIpDefault)
	viper.SetDefault(HttpIdleTimeoutKey, HttpIdleTimeoutDefault)
	viper.SetDefault(HttpReadTimeoutKey, HttpReadTimeoutDefault)
	viper.SetDefault(HttpWriteTimeoutKey, HttpWriteTimeoutDefault)
	viper.SetDefault(ClientIgnoreListByNameKey, ClientIgnoreListByNameDefault)
	viper.SetDefault(ClusterKey, ClusterDefault)
	viper.SetDefault(DbEnabledKey, DbEnabledDefault)
	viper.SetDefault(DbHostKey, DbHostDefault)
	viper.SetDefault(DbUsernameKey, DbUsernameDefault)
	viper.SetDefault(DbPasswordKey, DbPasswordDefault)
	viper.SetDefault(DbPortKey, DbPortDefault)
	viper.SetDefault(DbDatabaseKey, DbDatabaseDefault)
	viper.SetDefault(GhaDispatchEnabledKey, GhaDispatchEnabledDefault)
	viper.SetDefault(GhaTokenKey, GhaTokenDefault)
	viper.SetDefault(GhaUrlKey, GhaUrlDefault)
	viper.SetDefault(GhaOwnerKey, GhaOwnerDefault)
	viper.SetDefault(GhaRepoKey, GhaRepoDefault)
	viper.SetDefault(GhaEventTypeKey, GhaEventTypeDefault)
	viper.SetDefault(ExternalIntentsRetentionDaysKey, ExternalIntentsRetentionDaysDefault)

	excludedNamespaces = goset.FromSlice(viper.GetStringSlice(ExcludedNamespacesKey))
}
