package stage

import (
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/seancfoley/ipaddress-go/ipaddr"
)

// IpSet contains unique ip address
type IpSet struct {
	field          string
	ascendingOrder bool
	set            map[string]bool
	ips            []string
}

func NewIpSet(ascending bool, field string) *IpSet {

	return &IpSet{
		field:          field,
		ascendingOrder: ascending,
		set:            make(map[string]bool),
	}
}

func (p *IpSet) insert(key string) {

	if p.set[key] {
		return // Already in the map
	}
	p.ips = append(p.ips, key)
	p.set[key] = true
}

func (p *IpSet) Do(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg {
	var outMsgs []isplunk.SplunkPipeMsg

	if val, ok := input.Get("fields").(map[string]string); ok {
		outMsg := isplunk.NewSplunkMessage("unique.ips", nil)
		if ip, exists := val[p.field]; exists {
			address := ipaddr.NewIPAddressString(ip)
			if address.IsIPv4() {
				p.insert(ip)
				outMsg.Add("ipset", p.set)
				outMsg.Add("ips", p.ips)
				return append(outMsgs, outMsg)
			}
		}
	}

	return outMsgs
}
