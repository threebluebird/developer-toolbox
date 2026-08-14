package dnstool

import (
	"context"
	"developer-toolbox/backend/models"
	"net"
	"sort"
	"strings"
	"time"
)

type resolver interface {
	LookupHost(context.Context, string) ([]string, error)
	LookupCNAME(context.Context, string) (string, error)
	LookupMX(context.Context, string) ([]*net.MX, error)
	LookupTXT(context.Context, string) ([]string, error)
	LookupNS(context.Context, string) ([]*net.NS, error)
}
type DNSTool struct{ resolver resolver }

func NewDNSTool() *DNSTool { return &DNSTool{resolver: net.DefaultResolver} }
func (t *DNSTool) Info() models.Tool {
	return models.Tool{ID: "dns", Name: "DNS Lookup", Description: "Resolve A, AAAA, CNAME, MX, TXT, and NS records.", Category: "network", Icon: "dns", Version: "0.5.0", Capabilities: models.ToolCapabilities{Network: true}, Keywords: []string{"dns", "domain", "a", "aaaa", "cname", "mx", "txt", "ns"}}
}
func (t *DNSTool) Execute(tc models.ToolContext, input models.ToolInput) models.ToolOutput {
	host := strings.TrimSpace(stringValue(input.Payload["host"], stringValue(input.Payload["input"], "")))
	if host == "" {
		return bad("host required")
	}
	kind := strings.ToUpper(stringValue(input.Payload["recordType"], "A"))
	ctx, cancel := context.WithTimeout(tc.Context, 10*time.Second)
	defer cancel()
	values := []string{}
	switch kind {
	case "A", "AAAA":
		items, err := t.resolver.LookupHost(ctx, host)
		if err != nil {
			return bad(err.Error())
		}
		for _, item := range items {
			ip := net.ParseIP(item)
			if ip != nil && ((kind == "A" && ip.To4() != nil) || (kind == "AAAA" && ip.To4() == nil)) {
				values = append(values, item)
			}
		}
	case "CNAME":
		v, err := t.resolver.LookupCNAME(ctx, host)
		if err != nil {
			return bad(err.Error())
		}
		values = []string{v}
	case "MX":
		items, err := t.resolver.LookupMX(ctx, host)
		if err != nil {
			return bad(err.Error())
		}
		for _, v := range items {
			values = append(values, strings.TrimSuffix(v.Host, "."))
		}
	case "TXT":
		items, err := t.resolver.LookupTXT(ctx, host)
		if err != nil {
			return bad(err.Error())
		}
		values = items
	case "NS":
		items, err := t.resolver.LookupNS(ctx, host)
		if err != nil {
			return bad(err.Error())
		}
		for _, v := range items {
			values = append(values, strings.TrimSuffix(v.Host, "."))
		}
	default:
		return bad("unsupported record type")
	}
	sort.Strings(values)
	return models.ToolOutput{Success: true, Data: map[string]any{"host": host, "recordType": kind, "records": values}}
}
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return f
}
func bad(s string) models.ToolOutput { return models.Failure("DNS_ERROR", s) }
