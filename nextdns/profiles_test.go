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

func TestProfilesList(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": [
				{"id": "abc123", "fingerprint": "fp1", "name": "Profile 1"},
				{"id": "def456", "fingerprint": "fp2", "name": "Profile 2"}
			],
			"meta": {
				"pagination": {
					"cursor": ""
				}
			}
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	profiles, err := client.Profiles.List(ctx, &ListProfileRequest{})
	c.NoErr(err)

	c.Equal(len(profiles), 2)
	c.Equal(profiles[0].ID, "abc123")
	c.Equal(profiles[0].Fingerprint, "fp1")
	c.Equal(profiles[0].Name, "Profile 1")
	c.Equal(profiles[1].ID, "def456")
	c.Equal(profiles[1].Name, "Profile 2")
}

func TestProfilesCreate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPost)
		c.Equal(r.URL.Path, "/profiles")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":{"id":"new-profile-123"}}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreateProfileRequest{
		Name: "My New Profile",
		Security: &Security{
			ThreatIntelligenceFeeds: true,
			AiThreatDetection:       true,
			GoogleSafeBrowsing:      true,
			Cryptojacking:           true,
			DNSRebinding:            true,
			IdnHomographs:           true,
			Typosquatting:           true,
			Dga:                     false,
			Nrd:                     true,
			DDNS:                    false,
			Parking:                 true,
			Csam:                    true,
		},
		Privacy: &Privacy{
			DisguisedTrackers: true,
			AllowAffiliate:    false,
		},
		Denylist: []*Denylist{
			{ID: "malware.com", Active: true},
			{ID: "ads.example.com", Active: true},
		},
		Allowlist: []*Allowlist{
			{ID: "trusted.com", Active: true},
		},
		Settings: &Settings{
			Web3: true,
			Logs: &SettingsLogs{
				Enabled:   true,
				Retention: 30,
				Location:  "us",
			},
		},
	}

	id, err := client.Profiles.Create(ctx, request)
	c.NoErr(err)
	c.Equal(id, "new-profile-123")

	c.Equal(receivedBody["name"], "My New Profile")

	security, ok := receivedBody["security"].(map[string]interface{})
	c.True(ok)
	c.Equal(security["threatIntelligenceFeeds"], true)
	c.Equal(security["aiThreatDetection"], true)
	c.Equal(security["googleSafeBrowsing"], true)
	c.Equal(security["cryptojacking"], true)
	c.Equal(security["dnsRebinding"], true)
	c.Equal(security["idnHomographs"], true)
	c.Equal(security["typosquatting"], true)
	c.Equal(security["dga"], false)
	c.Equal(security["nrd"], true)
	c.Equal(security["ddns"], false)
	c.Equal(security["parking"], true)
	c.Equal(security["csam"], true)

	privacy, ok := receivedBody["privacy"].(map[string]interface{})
	c.True(ok)
	c.Equal(privacy["disguisedTrackers"], true)
	c.Equal(privacy["allowAffiliate"], false)

	denylist, ok := receivedBody["denylist"].([]interface{})
	c.True(ok)
	c.Equal(len(denylist), 2)

	settings, ok := receivedBody["settings"].(map[string]interface{})
	c.True(ok)
	c.Equal(settings["web3"], true)
}

func TestProfilesGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": {
				"name": "My Profile",
				"security": {
					"threatIntelligenceFeeds": true,
					"aiThreatDetection": true,
					"googleSafeBrowsing": true,
					"cryptojacking": true,
					"dnsRebinding": false,
					"idnHomographs": true,
					"typosquatting": true,
					"dga": false,
					"nrd": true,
					"ddns": false,
					"parking": true,
					"csam": true
				},
				"privacy": {
					"blocklists": [
						{"id": "nextdns-recommended"}
					],
					"natives": [
						{"id": "apple"}
					],
					"disguisedTrackers": true,
					"allowAffiliate": false
				},
				"parentalControl": {
					"services": [
						{"id": "tiktok", "active": true}
					],
					"categories": [
						{"id": "gambling", "active": true}
					],
					"safeSearch": true,
					"youtubeRestrictedMode": true,
					"blockBypass": false
				},
				"denylist": [
					{"id": "blocked.com", "active": true}
				],
				"allowlist": [
					{"id": "allowed.com", "active": true}
				],
				"settings": {
					"logs": {
						"enabled": true,
						"retention": 30,
						"location": "us"
					},
					"web3": true
				},
				"rewrites": [
					{"id": "rewrite-1", "name": "local.example.com", "type": "A", "content": "192.168.1.1"}
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
	profile, err := client.Profiles.Get(ctx, &GetProfileRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(profile.Name, "My Profile")

	c.True(profile.Security != nil)
	c.Equal(profile.Security.ThreatIntelligenceFeeds, true)
	c.Equal(profile.Security.AiThreatDetection, true)
	c.Equal(profile.Security.GoogleSafeBrowsing, true)
	c.Equal(profile.Security.DNSRebinding, false)
	c.Equal(profile.Security.Csam, true)

	c.True(profile.Privacy != nil)
	c.Equal(profile.Privacy.DisguisedTrackers, true)
	c.Equal(profile.Privacy.AllowAffiliate, false)
	c.Equal(len(profile.Privacy.Blocklists), 1)
	c.Equal(profile.Privacy.Blocklists[0].ID, "nextdns-recommended")

	c.True(profile.ParentalControl != nil)
	c.Equal(profile.ParentalControl.SafeSearch, true)
	c.Equal(profile.ParentalControl.YoutubeRestrictedMode, true)
	c.Equal(len(profile.ParentalControl.Services), 1)
	c.Equal(profile.ParentalControl.Services[0].ID, "tiktok")

	c.Equal(len(profile.Denylist), 1)
	c.Equal(profile.Denylist[0].ID, "blocked.com")

	c.Equal(len(profile.Allowlist), 1)
	c.Equal(profile.Allowlist[0].ID, "allowed.com")

	c.True(profile.Settings != nil)
	c.Equal(profile.Settings.Web3, true)
	c.True(profile.Settings.Logs != nil)
	c.Equal(profile.Settings.Logs.Enabled, true)
	c.Equal(profile.Settings.Logs.Retention, 30)

	c.Equal(len(profile.Rewrites), 1)
	c.Equal(profile.Rewrites[0].Name, "local.example.com")
	c.Equal(profile.Rewrites[0].Type, "A")
	c.Equal(profile.Rewrites[0].Content, "192.168.1.1")
}

func TestProfilesUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123")

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
	request := &UpdateProfileRequest{
		ProfileID: "abc123",
		Profile: &Profile{
			Name: "Updated Profile Name",
			Settings: &Settings{
				Web3: false,
			},
		},
	}

	err = client.Profiles.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["name"], "Updated Profile Name")
	settings, ok := receivedBody["settings"].(map[string]interface{})
	c.True(ok)
	c.Equal(settings["web3"], false)
}

func TestProfilesDelete(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodDelete)
		c.Equal(r.URL.Path, "/profiles/abc123")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	err = client.Profiles.Delete(ctx, &DeleteProfileRequest{ProfileID: "abc123"})
	c.NoErr(err)
}

func TestProfilesCreateWithParentalControlRecreation(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":{"id":"profile-with-recreation"}}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreateProfileRequest{
		Name: "Profile with Recreation",
		ParentalControl: &ParentalControl{
			SafeSearch:            true,
			YoutubeRestrictedMode: true,
			BlockBypass:           true,
			Recreation: &ParentalControlRecreation{
				Timezone: "America/New_York",
				Times: &ParentalControlRecreationTimes{
					Monday:    &ParentalControlRecreationInterval{Start: "15:00", End: "18:00"},
					Tuesday:   &ParentalControlRecreationInterval{Start: "15:00", End: "18:00"},
					Wednesday: &ParentalControlRecreationInterval{Start: "15:00", End: "18:00"},
					Thursday:  &ParentalControlRecreationInterval{Start: "15:00", End: "18:00"},
					Friday:    &ParentalControlRecreationInterval{Start: "15:00", End: "20:00"},
					Saturday:  &ParentalControlRecreationInterval{Start: "10:00", End: "22:00"},
					Sunday:    &ParentalControlRecreationInterval{Start: "10:00", End: "20:00"},
				},
			},
		},
	}

	id, err := client.Profiles.Create(ctx, request)
	c.NoErr(err)
	c.Equal(id, "profile-with-recreation")

	parentalControl, ok := receivedBody["parentalControl"].(map[string]interface{})
	c.True(ok)
	c.Equal(parentalControl["safeSearch"], true)
	c.Equal(parentalControl["youtubeRestrictedMode"], true)
	c.Equal(parentalControl["blockBypass"], true)

	recreation, ok := parentalControl["recreation"].(map[string]interface{})
	c.True(ok)
	c.Equal(recreation["timezone"], "America/New_York")

	times, ok := recreation["times"].(map[string]interface{})
	c.True(ok)

	monday, ok := times["monday"].(map[string]interface{})
	c.True(ok)
	c.Equal(monday["start"], "15:00")
	c.Equal(monday["end"], "18:00")

	saturday, ok := times["saturday"].(map[string]interface{})
	c.True(ok)
	c.Equal(saturday["start"], "10:00")
	c.Equal(saturday["end"], "22:00")
}
