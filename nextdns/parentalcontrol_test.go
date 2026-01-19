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

func TestParentalControlGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": {
				"services": [
					{"id": "tiktok", "active": true, "recreation": false},
					{"id": "instagram", "active": true, "recreation": true},
					{"id": "facebook", "active": false, "recreation": false}
				],
				"categories": [
					{"id": "gambling", "active": true, "recreation": false},
					{"id": "dating", "active": true, "recreation": false},
					{"id": "porn", "active": true, "recreation": false}
				],
				"recreation": {
					"timezone": "America/New_York",
					"times": {
						"monday": {"start": "16:00", "end": "20:00"},
						"tuesday": {"start": "16:00", "end": "20:00"},
						"wednesday": {"start": "16:00", "end": "20:00"},
						"thursday": {"start": "16:00", "end": "20:00"},
						"friday": {"start": "16:00", "end": "22:00"},
						"saturday": {"start": "09:00", "end": "22:00"},
						"sunday": {"start": "09:00", "end": "20:00"}
					}
				},
				"safeSearch": true,
				"youtubeRestrictedMode": true,
				"blockBypass": true
			}
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	pc, err := client.ParentalControl.Get(ctx, &GetParentalControlRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(pc.SafeSearch, true)
	c.Equal(pc.YoutubeRestrictedMode, true)
	c.Equal(pc.BlockBypass, true)

	c.Equal(len(pc.Services), 3)
	c.Equal(pc.Services[0].ID, "tiktok")
	c.Equal(pc.Services[0].Active, true)
	c.Equal(pc.Services[0].Recreation, false)
	c.Equal(pc.Services[1].ID, "instagram")
	c.Equal(pc.Services[1].Recreation, true)

	c.Equal(len(pc.Categories), 3)
	c.Equal(pc.Categories[0].ID, "gambling")
	c.Equal(pc.Categories[0].Active, true)

	c.True(pc.Recreation != nil)
	c.Equal(pc.Recreation.Timezone, "America/New_York")
	c.True(pc.Recreation.Times != nil)
	c.True(pc.Recreation.Times.Monday != nil)
	c.Equal(pc.Recreation.Times.Monday.Start, "16:00")
	c.Equal(pc.Recreation.Times.Monday.End, "20:00")
	c.Equal(pc.Recreation.Times.Saturday.Start, "09:00")
	c.Equal(pc.Recreation.Times.Saturday.End, "22:00")
}

func TestParentalControlUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl")

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
	request := &UpdateParentalControlRequest{
		ProfileID: "abc123",
		ParentalControl: &ParentalControl{
			SafeSearch:            true,
			YoutubeRestrictedMode: false,
			BlockBypass:           true,
			Recreation: &ParentalControlRecreation{
				Timezone: "Europe/London",
				Times: &ParentalControlRecreationTimes{
					Monday:   &ParentalControlRecreationInterval{Start: "17:00", End: "19:00"},
					Saturday: &ParentalControlRecreationInterval{Start: "10:00", End: "21:00"},
				},
			},
		},
	}

	err = client.ParentalControl.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["safeSearch"], true)
	c.Equal(receivedBody["youtubeRestrictedMode"], false)
	c.Equal(receivedBody["blockBypass"], true)

	recreation, ok := receivedBody["recreation"].(map[string]interface{})
	c.True(ok)
	c.Equal(recreation["timezone"], "Europe/London")

	times, ok := recreation["times"].(map[string]interface{})
	c.True(ok)

	monday, ok := times["monday"].(map[string]interface{})
	c.True(ok)
	c.Equal(monday["start"], "17:00")
	c.Equal(monday["end"], "19:00")
}

func TestParentalControlServicesList(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl/services")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": [
				{"id": "tiktok", "active": true, "recreation": false},
				{"id": "snapchat", "active": true, "recreation": true},
				{"id": "discord", "active": false, "recreation": false},
				{"id": "twitch", "active": true, "recreation": true}
			]
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	services, err := client.ParentalControlServices.List(ctx, &ListParentalControlServicesRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(len(services), 4)
	c.Equal(services[0].ID, "tiktok")
	c.Equal(services[0].Active, true)
	c.Equal(services[0].Recreation, false)
	c.Equal(services[1].ID, "snapchat")
	c.Equal(services[1].Recreation, true)
}

func TestParentalControlServicesCreate(t *testing.T) {
	c := is.New(t)

	var receivedBody []interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPut)
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl/services")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":[]}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreateParentalControlServicesRequest{
		ProfileID: "abc123",
		ParentalControlServices: []*ParentalControlServices{
			{ID: "tiktok", Active: true, Recreation: false},
			{ID: "youtube", Active: true, Recreation: true},
		},
	}

	err = client.ParentalControlServices.Create(ctx, request)
	c.NoErr(err)

	c.Equal(len(receivedBody), 2)

	first, ok := receivedBody[0].(map[string]interface{})
	c.True(ok)
	c.Equal(first["id"], "tiktok")
	c.Equal(first["active"], true)
	c.Equal(first["recreation"], false)
}

func TestParentalControlServicesUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl/services/tiktok")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":[]}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &UpdateParentalControlServicesRequest{
		ProfileID: "abc123",
		ID:        "tiktok",
		ParentalControlServices: &ParentalControlServices{
			Active:     false,
			Recreation: true,
		},
	}

	err = client.ParentalControlServices.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["active"], false)
	c.Equal(receivedBody["recreation"], true)
}

func TestParentalControlCategoriesList(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl/categories")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": [
				{"id": "gambling", "active": true, "recreation": false},
				{"id": "dating", "active": true, "recreation": false},
				{"id": "piracy", "active": true, "recreation": true},
				{"id": "social-networks", "active": false, "recreation": false}
			]
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	categories, err := client.ParentalControlCategories.List(ctx, &ListParentalControlCategoriesRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(len(categories), 4)
	c.Equal(categories[0].ID, "gambling")
	c.Equal(categories[0].Active, true)
	c.Equal(categories[2].ID, "piracy")
	c.Equal(categories[2].Recreation, true)
}

func TestParentalControlCategoriesCreate(t *testing.T) {
	c := is.New(t)

	var receivedBody []interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPut)
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl/categories")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":[]}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreateParentalControlCategoriesRequest{
		ProfileID: "abc123",
		ParentalControlCategories: []*ParentalControlCategories{
			{ID: "gambling", Active: true, Recreation: false},
			{ID: "porn", Active: true, Recreation: false},
		},
	}

	err = client.ParentalControlCategories.Create(ctx, request)
	c.NoErr(err)

	c.Equal(len(receivedBody), 2)

	first, ok := receivedBody[0].(map[string]interface{})
	c.True(ok)
	c.Equal(first["id"], "gambling")
	c.Equal(first["active"], true)
}

func TestParentalControlCategoriesUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl/categories/gambling")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":[]}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &UpdateParentalControlCategoriesRequest{
		ProfileID: "abc123",
		ID:        "gambling",
		ParentalControlCategories: &ParentalControlCategories{
			Active:     false,
			Recreation: false,
		},
	}

	err = client.ParentalControlCategories.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["active"], false)
	c.Equal(receivedBody["recreation"], false)
}

func TestParentalControlJSONFieldNames(t *testing.T) {
	c := is.New(t)

	pc := &ParentalControl{
		SafeSearch:            true,
		YoutubeRestrictedMode: true,
		BlockBypass:           true,
		Services: []*ParentalControlServices{
			{ID: "tiktok", Active: true, Recreation: false},
		},
		Categories: []*ParentalControlCategories{
			{ID: "gambling", Active: true, Recreation: false},
		},
		Recreation: &ParentalControlRecreation{
			Timezone: "UTC",
			Times: &ParentalControlRecreationTimes{
				Monday: &ParentalControlRecreationInterval{Start: "09:00", End: "17:00"},
			},
		},
	}

	data, err := json.Marshal(pc)
	c.NoErr(err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	c.NoErr(err)

	_, ok := result["safeSearch"]
	c.True(ok)
	_, ok = result["youtubeRestrictedMode"]
	c.True(ok)
	_, ok = result["blockBypass"]
	c.True(ok)
	_, ok = result["services"]
	c.True(ok)
	_, ok = result["categories"]
	c.True(ok)
	_, ok = result["recreation"]
	c.True(ok)

	recreation, ok := result["recreation"].(map[string]interface{})
	c.True(ok)
	_, ok = recreation["timezone"]
	c.True(ok)
	_, ok = recreation["times"]
	c.True(ok)

	times, ok := recreation["times"].(map[string]interface{})
	c.True(ok)
	monday, ok := times["monday"].(map[string]interface{})
	c.True(ok)
	_, ok = monday["start"]
	c.True(ok)
	_, ok = monday["end"]
	c.True(ok)
}
