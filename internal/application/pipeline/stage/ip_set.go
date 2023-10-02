package stage

import (
	"fmt"
	"net"

	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/speijnik/go-errortree"
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

func (p *IpSet) insert(key string) error {
	var rcerror error

	if p.set[key] {
		return errortree.Add(rcerror, "IpSet.insert", fmt.Errorf("%s already present", key)) // Already in the map
	}
	p.ips = append(p.ips, key)
	p.set[key] = true

	return nil
}

// only new ips are propagated down the pipeline
func (p *IpSet) Do(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg {
	var outMsgs []isplunk.SplunkPipeMsg

	if val, ok := input.Get("fields").(map[string]string); ok {
		outMsg := isplunk.NewSplunkMessage("unique.ips", nil)
		if ip, exists := val[p.field]; exists {
			if net.ParseIP(ip) != nil {
				if err := p.insert(ip); err != nil {
					return outMsgs
				}
				outMsg.Add("ipset", p.set)
				outMsg.Add("ips", p.ips)
				outMsg.Set("lastuniqueip", ip)
				return append(outMsgs, outMsg)
			}
		}
	}

	return outMsgs
}
