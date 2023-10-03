package stage

import (
	"bufio"
	"net"
	"os"

	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/speijnik/go-errortree"
	"github.com/yl2chen/cidranger"
)

const (
	ActiveSubnets   = "activecidr"   // subnets that contains a logged ip
	InactiveSubnets = "inactivecidr" // asubnets without matching logged ip
	ActiveIps       = "activeips"    // valid ips
	UnknonwIps      = "unknownips"   // ips that do not pertain to whitelisted subnets
)

// IpSet contains unique ip address
type MatchSubnet struct {
	whitelist       string
	file            *os.File
	ranger          cidranger.Ranger
	activesubnets   map[string]bool
	inactivesubnets map[string]bool
	activecidr      map[string]bool
	unknowncidr     map[string]bool
}

// openFile opens the file for reading.
func (m *MatchSubnet) openFile() error {
	var rcerror, err error

	if m.file, err = os.Open(m.whitelist); err != nil {
		return errortree.Add(rcerror, "matchsubnet.openFile", err)
	}

	return nil
}

func NewMatchSubnet(file string) (*MatchSubnet, error) {
	var rcerror, err error

	match := MatchSubnet{
		whitelist: file,
		ranger:    cidranger.NewPCTrieRanger(),
		//Whitelisted subnets used to connect
		activesubnets: make(map[string]bool),
		//Whitelisted subnets not used to connect
		inactivesubnets: make(map[string]bool),
		activecidr:      make(map[string]bool),
		// IPs not whitelisted
		unknowncidr: make(map[string]bool),
	}

	defer func() {
		if match.file != nil {
			match.file.Close()
		}
	}()

	if err = match.openFile(); err != nil {
		return nil, errortree.Add(rcerror, "NewMatchSubnet", err)
	}
	fileScanner := bufio.NewScanner(match.file)
	for tok := fileScanner.Scan(); tok; tok = fileScanner.Scan() {
		if _, network1, err := net.ParseCIDR(fileScanner.Text()); err != nil {
			return nil, errortree.Add(rcerror, "NewMatchSubnet", err)
		} else {
			address := cidranger.NewBasicRangerEntry(*network1)
			match.ranger.Insert(address)
			ipnet := address.Network()
			if !match.inactivesubnets[(&ipnet).String()] {
				match.inactivesubnets[(&ipnet).String()] = true
			}
		}
	}
	if err := fileScanner.Err(); err != nil {
		return nil, errortree.Add(rcerror, "NewMatchSubnet", err)
	}

	return &match, nil
}

// ip is sorted in two buckets known or unknown
func (m *MatchSubnet) Do(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg {
	var err error
	var outMsgs []isplunk.SplunkPipeMsg
	var contains bool

	if last, ok := input.Get(LastUniqueIpStageKey).(string); ok {
		outMsg := isplunk.NewSplunkMessage("match.subnets", nil)
		ip := net.ParseIP(last)
		if contains, err = m.ranger.Contains(ip); err != nil {
			return outMsgs
		}
		if contains {
			if nets, e := m.ranger.ContainingNetworks(ip); e != nil {
				return outMsgs
			} else {
				for _, cidr := range nets {
					p := cidr.Network()
					if m.inactivesubnets[(&p).String()] {
						delete(m.inactivesubnets, (&p).String())
						m.activesubnets[(&p).String()] = true
					}
				}
				if !m.activecidr[last] {
					m.activecidr[last] = true
				}
			}
		} else {
			if !m.unknowncidr[last] {
				m.unknowncidr[last] = true
			}
		}
		outMsg.Add(InactiveSubnets, m.inactivesubnets)
		outMsg.Add(ActiveSubnets, m.activesubnets)
		outMsg.Add(UnknonwIps, m.unknowncidr)
		outMsg.Add(ActiveIps, m.activecidr)

		return append(outMsgs, outMsg)
	}

	return outMsgs
}
