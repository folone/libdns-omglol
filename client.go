package omglol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

const apiBase = "https://api.omg.lol"

// listRecords fetches all DNS records for the configured address.
func (p *Provider) listRecords(ctx context.Context) ([]omglolRecord, error) {
	url := fmt.Sprintf("%s/address/%s/dns", apiBase, p.Address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("omg.lol API: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result omglolListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("omg.lol API: failed to parse response: %w", err)
	}

	if !result.Request.Success {
		return nil, fmt.Errorf("omg.lol API: request unsuccessful")
	}

	return result.Response.DNS, nil
}

// createRecord creates a new DNS record and returns the created record (with ID).
func (p *Provider) createRecord(ctx context.Context, payload omglolRecordPayload) (omglolRecord, error) {
	url := fmt.Sprintf("%s/address/%s/dns", apiBase, p.Address)

	data, err := json.Marshal(payload)
	if err != nil {
		return omglolRecord{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return omglolRecord{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return omglolRecord{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return omglolRecord{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return omglolRecord{}, fmt.Errorf("omg.lol API: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result omglolCreateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return omglolRecord{}, fmt.Errorf("omg.lol API: failed to parse create response: %w", err)
	}

	if !result.Request.Success {
		return omglolRecord{}, fmt.Errorf("omg.lol API: create unsuccessful")
	}

	return result.Response.ResponseReceived.Data, nil
}

// updateRecord updates an existing DNS record by its ID.
func (p *Provider) updateRecord(ctx context.Context, id string, payload omglolRecordPayload) error {
	url := fmt.Sprintf("%s/address/%s/dns/%s", apiBase, p.Address, id)

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("omg.lol API: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// deleteRecord deletes a DNS record by its ID.
func (p *Provider) deleteRecord(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/address/%s/dns/%s", apiBase, p.Address, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("omg.lol API: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// libdnsRecordToPayload converts a libdns.Record into the payload used by the
// omg.lol create/update endpoints. The name is converted from the libdns
// relative/absolute representation into what omg.lol expects: just the label
// prefix (e.g. "_acme-challenge" for "_acme-challenge.g.omg.lol.").
func libdnsRecordToPayload(record libdns.Record, zone string) omglolRecordPayload {
	rr := record.RR()
	ttl := int(rr.TTL / time.Second)
	if ttl <= 0 {
		ttl = 300
	}

	// Extract the address label from the zone (e.g. "g" from "g.omg.lol.").
	parts := strings.SplitN(strings.TrimSuffix(zone, "."), ".", 2)
	address := parts[0]

	// libdns names are relative to the zone or absolute (FQDN).
	// omg.lol expects just the sub-label portion relative to the address label.
	// For zone "g.omg.lol." the address is "g".
	// "_acme-challenge" relative to "g.omg.lol." → omg.lol name = "_acme-challenge"
	// "@" or "" (apex) → omg.lol name = address itself (e.g. "g")
	relativeName := libdns.RelativeName(rr.Name, zone)

	var omglolName string
	switch relativeName {
	case "@", "":
		omglolName = address
	default:
		// relativeName is e.g. "_acme-challenge" or "_acme-challenge.sub"
		omglolName = relativeName
	}

	return omglolRecordPayload{
		Type: rr.Type,
		Name: omglolName,
		Data: rr.Data,
		TTL:  ttl,
	}
}
