package omglol

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// omglolRecord represents a DNS record as returned by the omg.lol API.
type omglolRecord struct {
	ID       interface{} `json:"id"`
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	Data     string      `json:"data"`
	Priority *int        `json:"priority,omitempty"`
	TTL      interface{} `json:"ttl"`
}

// omglolListResponse is the top-level response for GET /address/{address}/dns.
type omglolListResponse struct {
	Request struct {
		StatusCode int  `json:"status_code"`
		Success    bool `json:"success"`
	} `json:"request"`
	Response struct {
		Message string         `json:"message"`
		DNS     []omglolRecord `json:"dns"`
	} `json:"response"`
}

// omglolCreateResponse is the response for POST /address/{address}/dns.
type omglolCreateResponse struct {
	Request struct {
		StatusCode int  `json:"status_code"`
		Success    bool `json:"success"`
	} `json:"request"`
	Response struct {
		Message          string `json:"message"`
		ResponseReceived struct {
			Data omglolRecord `json:"data"`
		} `json:"response_received"`
	} `json:"response"`
}

// omglolRecordPayload is the JSON body for create/update requests.
type omglolRecordPayload struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

// recordID extracts the record ID as a string regardless of whether the API
// returned it as a number or a string.
func (r omglolRecord) recordID() string {
	switch v := r.ID.(type) {
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ttlSeconds returns the TTL as an integer number of seconds.
func (r omglolRecord) ttlSeconds() int {
	switch v := r.TTL.(type) {
	case float64:
		return int(v)
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 3600
		}
		return n
	default:
		return 3600
	}
}

// toLibdnsRecord converts an omg.lol DNS record into a libdns.Record.
// The zone is the FQDN with trailing dot (e.g. "g.omg.lol.").
func (r omglolRecord) toLibdnsRecord(zone string) libdns.Record {
	ttl := time.Duration(r.ttlSeconds()) * time.Second

	// omg.lol returns the full name including the address label, e.g.
	// "g" or "_acme-challenge.g".  libdns wants a name relative to the zone.
	name := libdns.RelativeName(r.Name+"."+strings.TrimSuffix(zone, "."), zone)
	if name == "" {
		name = "@"
	}

	rr := libdns.RR{
		Name: name,
		TTL:  ttl,
		Type: r.Type,
		Data: r.Data,
	}

	return libdns.Record(rr)
}
