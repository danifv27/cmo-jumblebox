package stage

import (
	"net"

	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
)

// IpSet contains unique ip address
type MatchSubnet struct {
	whitelist string
	subnets   map[string]*net.IPNet
	known     map[string]bool
	unknown   map[string]bool
}

func NewMatchSubnet(file string) *MatchSubnet {

	return &MatchSubnet{
		whitelist: file,
		subnets:   make(map[string]*net.IPNet),
		known:     make(map[string]bool),
		unknown:   make(map[string]bool),
	}
}

// ip is sorted in two buckets known or unknown
func (p *MatchSubnet) Do(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg {
	var outMsgs []isplunk.SplunkPipeMsg

	if last, ok := input.Get("lastuniqueip").(string); ok {
		ip := net.ParseIP(last)
		outMsg := isplunk.NewSplunkMessage("match.subnets", nil)
		for _, subnet := range p.subnets {
			if subnet.Contains(ip) {
				if p.known[last] {
					p.known[last] = true
					outMsg.Add("knownset", p.known)
					break
				}
			} else {
				if p.unknown[last] {
					p.unknown[last] = true
					outMsg.Add("unknownset", p.unknown)
					break
				}
			}
		}
	}

	return outMsgs
}
