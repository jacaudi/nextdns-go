package nextdns

// LogStatus identifies the resolution status of a DNS query.
type LogStatus string

// LogStatus values returned by the NextDNS logs endpoint.
const (
	StatusDefault LogStatus = "default"
	StatusBlocked LogStatus = "blocked"
	StatusAllowed LogStatus = "allowed"
)

// String implements fmt.Stringer.
func (s LogStatus) String() string { return string(s) }

// SortOrder selects the ordering of log query results.
type SortOrder string

// SortOrder values accepted by the NextDNS logs endpoint.
const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// String implements fmt.Stringer.
func (s SortOrder) String() string { return string(s) }

// DestinationType selects the grouping for the destinations analytics endpoint.
type DestinationType string

// DestinationType values accepted by the NextDNS analytics destinations endpoint.
const (
	DestinationCountries DestinationType = "countries"
	DestinationGAFAM     DestinationType = "gafam"
)

// String implements fmt.Stringer.
func (d DestinationType) String() string { return string(d) }

// DNSProtocol identifies the wire protocol used for a DNS query.
type DNSProtocol string

// DNSProtocol values returned by the NextDNS analytics protocols endpoint.
const (
	ProtocolDoH DNSProtocol = "DNS-over-HTTPS"
	ProtocolDoT DNSProtocol = "DNS-over-TLS"
	ProtocolDoQ DNSProtocol = "DNS-over-QUIC"
	ProtocolUDP DNSProtocol = "UDP"
	ProtocolTCP DNSProtocol = "TCP"
)

// String implements fmt.Stringer.
func (p DNSProtocol) String() string { return string(p) }

// LogRetention is the configured log retention period, in seconds, as
// accepted by the NextDNS settings/logs endpoint.
type LogRetention int

// LogRetention values accepted by the NextDNS settings/logs endpoint, in seconds.
const (
	Retention1d   LogRetention = 86400
	Retention7d   LogRetention = 604800
	Retention30d  LogRetention = 2592000
	Retention90d  LogRetention = 7776000
	Retention180d LogRetention = 15552000
	Retention365d LogRetention = 31536000
)
