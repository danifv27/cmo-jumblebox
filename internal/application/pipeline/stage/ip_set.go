package stage

import (
	"fmt"
	"net"

	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/speijnik/go-errortree"
)

const (
	IPSetStageKey        = "ipset"        // map[string]bool with unique ips
	IPListStageKey       = "ips"          // array with IPSetStageKey map keys
	LastUniqueIpStageKey = "lastuniqueip" // last unique ip processed
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

	if val, ok := input.Get(ParsedFieldsStageKey).(map[string]string); ok {
		outMsg := isplunk.NewSplunkMessage("unique.ips", nil)
		if ip, exists := val[p.field]; exists {
			if net.ParseIP(ip) != nil {
				if err := p.insert(ip); err != nil {
					return outMsgs
				}
				outMsg.Add(IPSetStageKey, p.set)
				outMsg.Add(IPListStageKey, p.ips)
				outMsg.Set(LastUniqueIpStageKey, ip)
				return append(outMsgs, outMsg)
			}
		}
	}

	return outMsgs
}
