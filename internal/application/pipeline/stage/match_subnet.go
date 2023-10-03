package stage

import (
	"bufio"
	"net"
	"os"

	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/speijnik/go-errortree"
	"github.com/yl2chen/cidranger"
)

// IpSet contains unique ip address
type MatchSubnet struct {
	whitelist          string
	file               *os.File
	ranger             cidranger.Ranger
	usedsubnets        map[string]bool
	nonusedsubnets     map[string]bool
	whitelistedcidr    map[string]bool
	nonwhitelistedcidr map[string]bool
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
		usedsubnets: make(map[string]bool),
		//Whitelisted subnets not used to connect
		nonusedsubnets:  make(map[string]bool),
		whitelistedcidr: make(map[string]bool),
		// IPs not whitelisted
		nonwhitelistedcidr: make(map[string]bool),
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
			if !match.nonusedsubnets[(&ipnet).String()] {
				match.nonusedsubnets[(&ipnet).String()] = true
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

	if last, ok := input.Get("lastuniqueip").(string); ok {
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
					if m.nonusedsubnets[(&p).String()] {
						delete(m.nonusedsubnets, (&p).String())
						m.usedsubnets[(&p).String()] = true
					}
				}
				if !m.whitelistedcidr[last] {
					m.whitelistedcidr[last] = true
				}
			}
		} else {
			if !m.nonwhitelistedcidr[last] {
				m.nonwhitelistedcidr[last] = true
			}
		}
		outMsg.Add("nonusedsubnets", m.nonusedsubnets)
		outMsg.Add("usedsubnets", m.usedsubnets)
		outMsg.Add("nonwhitelistedips", m.nonwhitelistedcidr)
		outMsg.Add("whitelistedips", m.whitelistedcidr)

		return append(outMsgs, outMsg)
	}

	return outMsgs
}
