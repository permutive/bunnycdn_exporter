
CURL Command:

# load pull zones
curl --header "Content-Type: application/json"      --header "Accept: application/json"      --header "AccessKey: <API_KEY>"   'https://bunnycdn.com/api/pullzone' | jq

# load stats
curl --header "Content-Type: application/json"      --header "Accept: application/json"      --header "AccessKey: <API_KEY>"   'https://bunnycdn.com/api/statistics?dateFrom=2019-05-01&dateTo=2019-05-10&pullZone=65109&serverZoneId=-1&loadErrors=true' | jq

URL:
https://bunnycdn.com/api/statistics?
  dateFrom=2019-05-01&
  dateTo=2019-05-01&
  pullZone=65109&
  loadErrors=true




# this information is global (independent of pull zone)
"UserBalanceHistoryChart": {
  "2019-05-03T00:38:08": 1000
}
"UserStorageUsedChart": {
   "2019-05-01T06:30:29": 0
}

# then, for each pull zone we have to load stats again (setting the pullZoneId GET parameter)

## please totally ignore the summary at the top because they don't mean what you think it does ... it doesn't add up
"TotalBandwidthUsed": 82105090,
"TotalRequestsServed": 3774,
"CacheHitRate": 95.60148383677796,

# use the following then
"BandwidthUsedChart": {
  "2019-05-01T00:00:00Z": 32638120
},
"BandwidthCachedChart": {
  "2019-05-01T00:00:00Z": 32432597
},
"RequestsServedChart": {
  "2019-05-01T00:00:00Z": 1596
},
"PullRequestsPulledChart": {
  "2019-05-01T00:00:00Z": 166
},
"GeoTrafficDistribution": {
  "EU: London, UK": 13626096,
  "NA: Los Angeles, CA": 12778606,
  "NA: Atlanta, GA": 6410142,
  "NA: New York City, NY": 4814520,
  "EU: Amsterdam, NL": 5723172,
  "NA: Chicago, IL": 6410142,
  "EU: Oslo, NO": 6404216,
  "EU: Prague, CZ": 22724,
  "NA: Ashburn, VA": 1612474,
  "EU: Frankfurt, DE": 6337707
},
"Error3xxChart": {
  "2019-05-01T00:00:00Z": 1
},
"Error4xxChart": {
  "2019-05-01T00:00:00Z": 155
},
"Error5xxChart": {
  "2019-05-01T00:00:00Z": 4
}


# full Body (1 day)
{
  "TotalBandwidthUsed": 64139799,
  "TotalRequestsServed": 2983,
  "CacheHitRate": 94.43513241702983,
  "BandwidthUsedChart": {
    "2019-05-01T00:00:00Z": 32638120
  },
  "BandwidthCachedChart": {
    "2019-05-01T00:00:00Z": 32432597
  },
  "CacheHitRateChart": {
    "2019-05-01T00:00:00Z": 0
  },
  "RequestsServedChart": {
    "2019-05-01T00:00:00Z": 1596
  },
  "PullRequestsPulledChart": {
    "2019-05-01T00:00:00Z": 166
  },
  "UserBalanceHistoryChart": {
    "2019-05-01T00:37:25": 1000
  },
  "UserStorageUsedChart": {
    "2019-05-01T06:30:29": 0
  },
  "GeoTrafficDistribution": {
    "EU: London, UK": 13626096,
    "NA: Los Angeles, CA": 12778606,
    "NA: Atlanta, GA": 6410142,
    "NA: New York City, NY": 4814520,
    "EU: Amsterdam, NL": 5723172,
    "NA: Chicago, IL": 6410142,
    "EU: Oslo, NO": 6404216,
    "EU: Prague, CZ": 22724,
    "NA: Ashburn, VA": 1612474,
    "EU: Frankfurt, DE": 6337707
  },
  "Error3xxChart": {
    "2019-05-01T00:00:00Z": 1
  },
  "Error4xxChart": {
    "2019-05-01T00:00:00Z": 155
  },
  "Error5xxChart": {
    "2019-05-01T00:00:00Z": 4
  }
}

# full body (over 3 days)
{
  "TotalBandwidthUsed": 83422382,
  "TotalRequestsServed": 3832,
  "CacheHitRate": 95.66805845511482,
  "BandwidthUsedChart": {
    "2019-05-01T00:00:00Z": 32638120,
    "2019-05-02T00:00:00Z": 32750843,
    "2019-05-03T00:00:00Z": 18033419
  },
  "BandwidthCachedChart": {
    "2019-05-01T00:00:00Z": 32432597,
    "2019-05-02T00:00:00Z": 32750843,
    "2019-05-03T00:00:00Z": 18033419
  },
  "CacheHitRateChart": {
    "2019-05-01T00:00:00Z": 0,
    "2019-05-02T00:00:00Z": 0,
    "2019-05-03T00:00:00Z": 0
  },
  "RequestsServedChart": {
    "2019-05-01T00:00:00Z": 1596,
    "2019-05-02T00:00:00Z": 1442,
    "2019-05-03T00:00:00Z": 794
  },
  "PullRequestsPulledChart": {
    "2019-05-01T00:00:00Z": 166,
    "2019-05-02T00:00:00Z": 0,
    "2019-05-03T00:00:00Z": 0
  },
  "UserBalanceHistoryChart": {
    "2019-05-01T00:37:25": 1000,
    "2019-05-02T00:37:51": 1000,
    "2019-05-03T00:38:08": 1000
  },
  "UserStorageUsedChart": {
    "2019-05-01T06:30:29": 0,
    "2019-05-02T09:35:02": 0,
    "2019-05-03T06:34:31": 0
  },
  "GeoTrafficDistribution": {
    "EU: London, UK": 17600346,
    "NA: Los Angeles, CA": 16636814,
    "NA: Atlanta, GA": 8365008,
    "NA: New York City, NY": 6744870,
    "EU: Amsterdam, NL": 7517341,
    "NA: Chicago, IL": 8342277,
    "EU: Oslo, NO": 7085516,
    "EU: Prague, CZ": 22724,
    "NA: Ashburn, VA": 2861579,
    "EU: Frankfurt, DE": 8245907
  },
  "Error3xxChart": {
    "2019-05-01T00:00:00Z": 1,
    "2019-05-02T00:00:00Z": 0,
    "2019-05-03T00:00:00Z": 0
  },
  "Error4xxChart": {
    "2019-05-01T00:00:00Z": 155,
    "2019-05-02T00:00:00Z": 0,
    "2019-05-03T00:00:00Z": 0
  },
  "Error5xxChart": {
    "2019-05-01T00:00:00Z": 4,
    "2019-05-02T00:00:00Z": 0,
    "2019-05-03T00:00:00Z": 0
  }
}
