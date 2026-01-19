package nextdns

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestSecurityGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/security")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": {
				"threatIntelligenceFeeds": true,
				"aiThreatDetection": true,
				"googleSafeBrowsing": true,
				"cryptojacking": true,
				"dnsRebinding": true,
				"idnHomographs": true,
				"typosquatting": true,
				"dga": false,
				"nrd": true,
				"ddns": false,
				"parking": true,
				"csam": true,
				"tlds": [
					{"id": ".xyz"},
					{"id": ".top"}
				]
			}
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	security, err := client.Security.Get(ctx, &GetSecurityRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(security.ThreatIntelligenceFeeds, true)
	c.Equal(security.AiThreatDetection, true)
	c.Equal(security.GoogleSafeBrowsing, true)
	c.Equal(security.Cryptojacking, true)
	c.Equal(security.DNSRebinding, true)
	c.Equal(security.IdnHomographs, true)
	c.Equal(security.Typosquatting, true)
	c.Equal(security.Dga, false)
	c.Equal(security.Nrd, true)
	c.Equal(security.DDNS, false)
	c.Equal(security.Parking, true)
	c.Equal(security.Csam, true)

	c.Equal(len(security.Tlds), 2)
	c.Equal(security.Tlds[0].ID, ".xyz")
	c.Equal(security.Tlds[1].ID, ".top")
}

func TestSecurityUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/security")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":{}}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &UpdateSecurityRequest{
		ProfileID: "abc123",
		Security: &Security{
			ThreatIntelligenceFeeds: true,
			AiThreatDetection:       true,
			GoogleSafeBrowsing:      true,
			Cryptojacking:           false,
			DNSRebinding:            true,
			IdnHomographs:           true,
			Typosquatting:           true,
			Dga:                     true,
			Nrd:                     false,
			DDNS:                    true,
			Parking:                 false,
			Csam:                    true,
		},
	}

	err = client.Security.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["threatIntelligenceFeeds"], true)
	c.Equal(receivedBody["aiThreatDetection"], true)
	c.Equal(receivedBody["googleSafeBrowsing"], true)
	c.Equal(receivedBody["cryptojacking"], false)
	c.Equal(receivedBody["dnsRebinding"], true)
	c.Equal(receivedBody["idnHomographs"], true)
	c.Equal(receivedBody["typosquatting"], true)
	c.Equal(receivedBody["dga"], true)
	c.Equal(receivedBody["nrd"], false)
	c.Equal(receivedBody["ddns"], true)
	c.Equal(receivedBody["parking"], false)
	c.Equal(receivedBody["csam"], true)
}

func TestSecurityJSONFieldNames(t *testing.T) {
	c := is.New(t)

	security := &Security{
		ThreatIntelligenceFeeds: true,
		AiThreatDetection:       true,
		GoogleSafeBrowsing:      true,
		Cryptojacking:           true,
		DNSRebinding:            true,
		IdnHomographs:           true,
		Typosquatting:           true,
		Dga:                     true,
		Nrd:                     true,
		DDNS:                    true,
		Parking:                 true,
		Csam:                    true,
		Tlds: []*SecurityTlds{
			{ID: ".xyz"},
		},
	}

	data, err := json.Marshal(security)
	c.NoErr(err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	c.NoErr(err)

	_, ok := result["threatIntelligenceFeeds"]
	c.True(ok)
	_, ok = result["aiThreatDetection"]
	c.True(ok)
	_, ok = result["googleSafeBrowsing"]
	c.True(ok)
	_, ok = result["cryptojacking"]
	c.True(ok)
	_, ok = result["dnsRebinding"]
	c.True(ok)
	_, ok = result["idnHomographs"]
	c.True(ok)
	_, ok = result["typosquatting"]
	c.True(ok)
	_, ok = result["dga"]
	c.True(ok)
	_, ok = result["nrd"]
	c.True(ok)
	_, ok = result["ddns"]
	c.True(ok)
	_, ok = result["parking"]
	c.True(ok)
	_, ok = result["csam"]
	c.True(ok)
	_, ok = result["tlds"]
	c.True(ok)
}
