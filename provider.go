// Package omglol implements a DNS record management client compatible
// with the libdns interfaces for omg.lol.
package omglol

import (
	"context"
	"strings"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with omg.lol.
//
// Address is the omg.lol address (handle) that owns the zone, e.g. "g" for
// the zone "g.omg.lol.".  APIKey is the omg.lol API key used for
// authentication.
type Provider struct {
	// APIKey is your omg.lol API key.
	APIKey string `json:"api_key,omitempty"`

	// Address is your omg.lol address/handle (e.g. "yourname" for yourname.omg.lol).
	Address string `json:"address,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	raw, err := p.listRecords(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]libdns.Record, 0, len(raw))
	for _, r := range raw {
		records = append(records, r.toLibdnsRecord(zone))
	}
	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var created []libdns.Record

	for _, record := range records {
		payload := libdnsRecordToPayload(record, zone)

		result, err := p.createRecord(ctx, payload)
		if err != nil {
			return created, err
		}

		created = append(created, result.toLibdnsRecord(zone))
	}

	return created, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones.  It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	existing, err := p.listRecords(ctx)
	if err != nil {
		return nil, err
	}

	var results []libdns.Record

	for _, record := range records {
		rr := record.RR()
		payload := libdnsRecordToPayload(record, zone)

		// Look for an existing record with the same type and name.
		var matched *omglolRecord
		for i := range existing {
			e := &existing[i]
			if strings.EqualFold(e.Type, rr.Type) && namesMatch(e.Name, payload.Name, zone) {
				matched = e
				break
			}
		}

		if matched != nil {
			id := matched.recordID()
			if err := p.updateRecord(ctx, id, payload); err != nil {
				return results, err
			}
			// Reflect updated values back as a libdns record.
			updated := omglolRecord{
				ID:   matched.ID,
				Type: payload.Type,
				Name: payload.Name,
				Data: payload.Data,
				TTL:  payload.TTL,
			}
			results = append(results, updated.toLibdnsRecord(zone))
		} else {
			created, err := p.createRecord(ctx, payload)
			if err != nil {
				return results, err
			}
			results = append(results, created.toLibdnsRecord(zone))
		}
	}

	return results, nil
}

// DeleteRecords deletes the records from the zone.  It returns the records
// that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	existing, err := p.listRecords(ctx)
	if err != nil {
		return nil, err
	}

	var deleted []libdns.Record

	for _, record := range records {
		rr := record.RR()
		payload := libdnsRecordToPayload(record, zone)

		for _, e := range existing {
			if !strings.EqualFold(e.Type, rr.Type) {
				continue
			}
			if !namesMatch(e.Name, payload.Name, zone) {
				continue
			}
			// If Data is specified, also match on content.
			if rr.Data != "" && e.Data != rr.Data {
				continue
			}

			id := e.recordID()
			if err := p.deleteRecord(ctx, id); err != nil {
				return deleted, err
			}
			deleted = append(deleted, e.toLibdnsRecord(zone))
		}
	}

	return deleted, nil
}

// namesMatch compares the name returned by the omg.lol API (e.g. "_acme-challenge.g")
// with the name derived from a libdns record (e.g. "_acme-challenge").
//
// omg.lol stores names as "<label>.<address>" (e.g. "_acme-challenge.g") or
// just "<address>" for the apex record.  The payload name we build is the
// label portion only (e.g. "_acme-challenge" or "g" for apex).
func namesMatch(apiName, payloadName, zone string) bool {
	// Normalise: strip trailing dot from zone and extract address label.
	zoneNoSuffix := strings.TrimSuffix(zone, ".")
	parts := strings.SplitN(zoneNoSuffix, ".", 2)
	address := parts[0] // e.g. "g"

	// apiName examples: "g", "_acme-challenge.g", "sub._acme-challenge.g"
	// payloadName examples: "g", "_acme-challenge"

	// Strip the ".<address>" suffix from the API name to get the raw label.
	apiLabel := strings.TrimSuffix(apiName, "."+address)
	if apiLabel == address {
		// apex record
		apiLabel = address
	}

	return strings.EqualFold(apiLabel, payloadName) || strings.EqualFold(apiName, payloadName)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
