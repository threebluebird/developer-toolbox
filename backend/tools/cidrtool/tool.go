package cidrtool

import (
	"developer-toolbox/backend/models"
	"math"
	"net"
	"strconv"
	"strings"
)

type CIDRTool struct{}

func NewCIDRTool() *CIDRTool { return &CIDRTool{} }
func (t *CIDRTool) Info() models.Tool {
	return models.Tool{ID: "cidr", Name: "CIDR Calculator", Description: "Calculate IPv4 and IPv6 network ranges.", Category: "network", Icon: "cidr", Version: "0.5.0", Keywords: []string{"cidr", "subnet", "network", "broadcast", "mask", "ip"}}
}
func (t *CIDRTool) Execute(input models.ToolInput) models.ToolOutput {
	raw := strings.TrimSpace(stringValue(input.Payload["input"], ""))
	ip, network, err := net.ParseCIDR(raw)
	if err != nil {
		return bad(err.Error())
	}
	ones, bits := network.Mask.Size()
	first := network.IP
	last := make(net.IP, len(first))
	for i := range first {
		last[i] = first[i] | ^network.Mask[i]
	}
	result := map[string]any{"input": ip.String(), "network": network.String(), "firstIP": first.String(), "lastIP": last.String(), "prefix": ones}
	if bits == 32 {
		total := uint64(1) << uint(bits-ones)
		result["subnetMask"] = net.IP(network.Mask).String()
		result["broadcast"] = last.String()
		result["totalHosts"] = total
		if total > 2 {
			usable := total - 2
			result["usableHosts"] = usable
			result["firstUsable"] = increment(first).String()
			result["lastUsable"] = decrement(last).String()
		} else {
			result["usableHosts"] = uint64(0)
		}
	} else {
		hostBits := bits - ones
		if hostBits <= 63 {
			result["totalHosts"] = uint64(math.Pow(2, float64(hostBits)))
		} else {
			result["totalHosts"] = "2^" + strconv.Itoa(hostBits)
		}
	}
	return models.ToolOutput{Success: true, Data: result}
}
func increment(ip net.IP) net.IP {
	v := append(net.IP(nil), ip...)
	for i := len(v) - 1; i >= 0; i-- {
		v[i]++
		if v[i] != 0 {
			break
		}
	}
	return v
}
func decrement(ip net.IP) net.IP {
	v := append(net.IP(nil), ip...)
	for i := len(v) - 1; i >= 0; i-- {
		before := v[i]
		v[i]--
		if before != 0 {
			break
		}
	}
	return v
}
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return f
}
func bad(s string) models.ToolOutput { return models.Failure("CIDR_ERROR", s) }
