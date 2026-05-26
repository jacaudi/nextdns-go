package nextdns

import (
	"testing"

	"github.com/matryer/is"
)

func TestLogStatusConstants(t *testing.T) {
	c := is.New(t)
	c.Equal(string(StatusDefault), "default")
	c.Equal(string(StatusBlocked), "blocked")
	c.Equal(string(StatusAllowed), "allowed")
	c.Equal(string(StatusError), "error")
}

func TestSortOrderConstants(t *testing.T) {
	c := is.New(t)
	c.Equal(string(SortAsc), "asc")
	c.Equal(string(SortDesc), "desc")
}

func TestDestinationTypeConstants(t *testing.T) {
	c := is.New(t)
	c.Equal(string(DestinationCountries), "countries")
	c.Equal(string(DestinationGAFAM), "gafam")
}

func TestDNSProtocolConstants(t *testing.T) {
	c := is.New(t)
	c.Equal(string(ProtocolDoH), "DNS-over-HTTPS")
	c.Equal(string(ProtocolDoT), "DNS-over-TLS")
	c.Equal(string(ProtocolDoQ), "DNS-over-QUIC")
	c.Equal(string(ProtocolUDP), "UDP")
	c.Equal(string(ProtocolTCP), "TCP")
}

func TestLogRetentionConstants(t *testing.T) {
	c := is.New(t)
	c.Equal(int(Retention1d), 86400)
	c.Equal(int(Retention7d), 604800)
	c.Equal(int(Retention30d), 2592000)
	c.Equal(int(Retention90d), 7776000)
	c.Equal(int(Retention180d), 15552000)
	c.Equal(int(Retention365d), 31536000)
}
