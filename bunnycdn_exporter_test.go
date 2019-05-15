// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testSocket = "/tmp/bunnycdnexportertest.sock"

type bunny struct {
	*httptest.Server
	responsePullZones []byte
	responseStats     []byte
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a != b {
		t.Fatalf("%s | expected: %v != found: %v", message, a, b)
	}
}

func newBunny(responsePullZones []byte, responseStats []byte) *bunny {
	h := &bunny{responsePullZones: responsePullZones, responseStats: responseStats}
	h.Server = httptest.NewServer(handler(h))
	return h
}

func handler(h *bunny) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pullzone" {
			w.Write(h.responsePullZones)
		} else if r.URL.Path == "/statistics" {
			w.Write(h.responseStats)
		} else {
			w.Write([]byte("error"))
		}
	}
}

func handlerStale(exit chan bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		<-exit
	}
}

func TestPullZoneList(t *testing.T) {
	respBody := []byte(`[{"Id": 34567,"Name": "pullzonename2","OriginUrl": "https://storage.googleapis.com/gcs-bucket-example","Enabled": true,"Hostnames": [{"Id": 43217,"Value": "pullzonename2.b-cdn.net","ForceSSL": true,"IsSystemHostname": true,"HasCertificate": true}],"StorageZoneId": 0,"AllowedReferrers": [],"BlockedReferrers": [],"BlockedIps": [],"EnableGeoZoneUS": true,"EnableGeoZoneEU": true,"EnableGeoZoneASIA": true,"EnableGeoZoneSA": true,"EnableGeoZoneAF": true,"ZoneSecurityEnabled": false,"ZoneSecurityKey": "72730e8f-08ff-4db3-8fa7-965097a1755f","ZoneSecurityIncludeHashRemoteIP": false,"IgnoreQueryStrings": true,"MonthlyBandwidthLimit": 0,"MonthlyBandwidthUsed": 74309996,"MonthlyCharges": 0.000831675729999988,"AddHostHeader": false,"Type": 0,"CustomNginxConfig": "","AccessControlOriginHeaderExtensions": ["eot","ttf","woff","woff2","css"],"EnableAccessControlOriginHeader": true,"DisableCookies": true,"BudgetRedirectedCountries": [],"BlockedCountries": [],"EnableOriginShield": false,"CacheControlMaxAgeOverride": -1,"BurstSize": 0,"RequestLimit": 0,"BlockRootPathAccess": false,"CacheQuality": 75,"LimitRatePerSecond": 0,"LimitRateAfter": 0,"ConnectionLimitPerIPCount": 0,"PriceOverride": 0,"AddCanonicalHeader": false,"EnableLogging": true,"IgnoreVaryHeader": true,"EnableCacheSlice": false,"EdgeRules": [{"Guid": "82730e8f-08ff-4db3-8fa7-965097a1755f","ActionType": 2,"ActionParameter1": "https://storage.googleapis.com/gcs-bucket/folder/","ActionParameter2": "","Triggers": [{"Type": 0,"PatternMatches": ["https://pullzonename2.b-cdn.net/folder/*"],"PatternMatchingType": 0,"Parameter1": ""}],"TriggerMatchingType": 0,"Description": "models","Enabled": true}],"EnableWebPVary": false,"EnableCountryCodeVary": false,"EnableMobileVary": false,"EnableHostnameVary": false,"CnameDomain": "b-cdn.net"},{"Id": 12345,"Name": "pullzonename","OriginUrl": "http://origin.url.com","Enabled": true,"Hostnames": [{"Id": 54321,"Value": "pullzonename.b-cdn.net","ForceSSL": false,"IsSystemHostname": true,"HasCertificate": true}],"StorageZoneId": 0,"AllowedReferrers": [],"BlockedReferrers": [],"BlockedIps": [],"EnableGeoZoneUS": true,"EnableGeoZoneEU": true,"EnableGeoZoneASIA": true,"EnableGeoZoneSA": true,"EnableGeoZoneAF": true,"ZoneSecurityEnabled": false,"ZoneSecurityKey": "92730e8f-08ff-4db3-8fa7-965097a1755f","ZoneSecurityIncludeHashRemoteIP": false,"IgnoreQueryStrings": true,"MonthlyBandwidthLimit": 0,"MonthlyBandwidthUsed": 0,"MonthlyCharges": 0,"AddHostHeader": false,"Type": 0,"CustomNginxConfig": "","AccessControlOriginHeaderExtensions": ["eot","ttf","woff","woff2","css"],"EnableAccessControlOriginHeader": true,"DisableCookies": true,"BudgetRedirectedCountries": [],"BlockedCountries": [],"EnableOriginShield": false,"CacheControlMaxAgeOverride": -1,"BurstSize": 0,"RequestLimit": 0,"BlockRootPathAccess": false,"CacheQuality": 75,"LimitRatePerSecond": 0,"LimitRateAfter": 0,"ConnectionLimitPerIPCount": 0,"PriceOverride": 0,"AddCanonicalHeader": false,"EnableLogging": true,"IgnoreVaryHeader": true,"EnableCacheSlice": false,"EdgeRules": [],"EnableWebPVary": false,"EnableCountryCodeVary": false,"EnableMobileVary": false,"EnableHostnameVary": false,"CnameDomain": "b-cdn.net"}]`)

	h := newBunny(respBody, respBody)

	fetch := fetchHTTP(h.URL, "api_key", true, time.Second*1)
	pullZones, err := listPullZones(fetch)
	if err != nil {
		t.Fatal("Unexpected error listing pull zones: ", err)
	}
	if len(pullZones) != 2 {
		t.Fatal("Expecting 2 zones but got ", len(pullZones))
	}
	if pullZones[0].ID != 34567 || pullZones[0].Name != "pullzonename2" {
		t.Fatal("Expecting ID 34567 and name pullzonename2 but got " + string(pullZones[0].ID) + " and " + pullZones[0].Name)
	}
}

func TestStatistics(t *testing.T) {
	respBody := []byte(`{"TotalBandwidthUsed": 28639956,"TotalRequestsServed": 1261,"CacheHitRate": 100,"BandwidthUsedChart": {"2019-05-02T00:00:00Z": 28639956},"BandwidthCachedChart": {"2019-05-02T00:00:00Z": 28639956},"CacheHitRateChart": {"2019-05-02T00:00:00Z": 0},"RequestsServedChart": {"2019-05-02T00:00:00Z": 1261},"PullRequestsPulledChart": {"2019-05-02T00:00:00Z": 0},"UserBalanceHistoryChart": {"2019-05-02T00:37:51": 1000},"UserStorageUsedChart": {"2019-05-02T09:35:02": 0},"GeoTrafficDistribution": {"EU: London, UK": 6040860,"NA: Los Angeles, CA": 5719265,"NA: Atlanta, GA": 2864106,"NA: New York City, NY": 2861460,"EU: Amsterdam, NL": 2566343,"NA: Chicago, IL": 2864106,"EU: Oslo, NO": 2884170,"EU: Frankfurt, DE": 2839646},"Error3xxChart": {"2019-05-02T00:00:00Z": 0},"Error4xxChart": {"2019-05-02T00:00:00Z": 0},"Error5xxChart": {"2019-05-02T00:00:00Z": 0}}`)

	h := newBunny(respBody, respBody)

	fetch := fetchHTTP(h.URL, "api_key", true, time.Second*1)
	stats, err := getStatistics(fetch)
	if err != nil {
		t.Fatal("Unexpected error getting stats: ", err)
	}
	assertEqual(t, float64(28639956), extractFromMap(stats.BandwidthUsed), "Unexpected bandwidth used total.")
}

func TestTrafficLocationSplit(t *testing.T) {
	respBody := []byte(`{"GeoTrafficDistribution": {"EU: London, UK": 6040860.0}}`)

	h := newBunny(respBody, respBody)

	fetch := fetchHTTP(h.URL, "api_key", true, time.Second*1)
	stats, err := getStatistics(fetch)
	if err != nil {
		t.Fatal("Unexpected error getting stats for testing traffic location: ", err)
	}
	locs := stats.trafficLocations()

	if len(locs) != 1 {
		t.Fatal("Number of locations found: expected: 1, got: ", len(locs))
	}

	assertEqual(t, "EU", locs[0].Region, "Region for location")
	assertEqual(t, "London, UK", locs[0].Location, "Location description")
	assertEqual(t, float64(6040860), locs[0].Requests, "Number of requests for location")
}
