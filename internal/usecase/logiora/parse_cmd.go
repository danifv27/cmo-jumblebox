package logiora

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	apipe "fry.org/qft/jumble/internal/application/pipeline"
	"fry.org/qft/jumble/internal/application/pipeline/stage"
	"fry.org/qft/jumble/internal/application/printer"
	ifollower "fry.org/qft/jumble/internal/infrastructure/follower/file"
	ipipe "fry.org/qft/jumble/internal/infrastructure/pipeline"
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/speijnik/go-errortree"
	"github.com/tidwall/pretty"
)

type ParseCmd struct {
	// Format string `kong:"help='Log format',default='$remote_addr - $remote_user [$time_local] \"$request\" $status $body_bytes_sent \"$http_referer\" \"$http_user_agent\" rt=$request_time uct=\"$upstream_connect_time\" uht=\"$upstream_header_time\" urt=\"$upstream_response_time\" x_request-id=\"$http_x_request_id\" reqid=\"$reqid\"'"`
	Format string `kong:"help='Log format',default='$http_x_original_forwarded_for - $remote_addr - $remote_user [$time_local] - $request $status'"`
	File   struct {
		File  string   `kong:"arg,help='File to parse'"`
		Match MatchCmd `kong:"cmd,help='Check if ip is whitelisted'"`
	} `kong:"arg"`
}

type MatchCmd struct {
	Whitelist struct {
		Whitelist string `kong:"arg,help='Whitelisted ranges'"`
	} `kong:"arg"`
	lastMsg isplunk.SplunkPipeMsg
}

func do(ctx context.Context, cancel context.CancelFunc, pipe apipe.Piper[isplunk.SplunkPipeMsg], entries chan string, errs chan error) (<-chan isplunk.SplunkPipeMsg, error) {
	var rcerror, err error
	var outCh <-chan isplunk.SplunkPipeMsg

	inCh := make(chan isplunk.SplunkPipeMsg)
	go func(ct context.Context) {
		defer func() {
			close(inCh)
		}()
		for {
			select {
			case entry, hasMore := <-entries:
				if !hasMore {
					// fmt.Println("[DBG]No more entries")
					return
				}
				expected := isplunk.NewSplunkMessage("input.entry", nil)
				expected.Add("entry", entry)
				inCh <- expected
			case failure, hasMore := <-errs:
				if !hasMore || failure != nil {
					// fmt.Println("[DBG]No more errors")
					return
				}
			case <-ctx.Done():
				fmt.Printf("[DBG]context cancelled. Stopping source goroutine\n")
				return
			}
		}
	}(ctx)

	// Run the pipeline
	if outCh, err = pipe.Do(ctx, inCh); err != nil {
		fmt.Printf("[DBG]pipe.Do err: %s\n", err)
		cancel()
		return outCh, errortree.Add(rcerror, "matchcmd.do", err)
	}

	return outCh, nil
}

func (cmd *MatchCmd) Run(cli *CLI) error {
	var rcerror, err error
	var ppln apipe.Piper[isplunk.SplunkPipeMsg]
	var outCh <-chan isplunk.SplunkPipeMsg

	//The source of the pipeline are the lines from the log file
	follow := ifollower.NewFollower(cli.Parse.File.File)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	//Prepare the pipeline
	if ppln, err = ipipe.Parse[isplunk.SplunkPipeMsg]("pipeline:splunk"); err != nil {
		return errortree.Add(rcerror, "matchcmd.do", err)
	}
	// Pipeline stages
	// Regexp parse
	regexParserStg := stage.NewRegexParse(cli.Parse.Format)
	st := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](regexParserStg.Do, isplunk.WithName("regexParser"), isplunk.WithBufferSize(0))
	ppln.Next(st)
	// Check unique ip
	ipStg := stage.NewIpSet(true, "http_x_original_forwarded_for")
	ipset := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](ipStg.Do, isplunk.WithName("ipSet"))
	ppln.Next(ipset)
	// Match subnets
	if matchStg, err := stage.NewMatchSubnet(cli.Parse.File.Match.Whitelist.Whitelist); err != nil {
		return errortree.Add(rcerror, "matchcmd.do", err)
	} else {
		matchsubnet := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](matchStg.Do, isplunk.WithName("matchSubnet"))
		ppln.Next(matchsubnet)
	}

	// Pipeline source
	entriesCh, errorsCh, err := follow.Lines(ctx)
	if err != nil {
		cancel()
		return errortree.Add(rcerror, "matchcmd.Run", err)
	}

	if outCh, err = do(ctx, cancel, ppln, entriesCh, errorsCh); err != nil {
		cancel()
		return errortree.Add(rcerror, "matchcmd.Run", err)
	}

	// Drain pipeline
	count := 0
mainLoop:
	for {
		select {
		case msg, more := <-outCh:
			if !more {
				fmt.Printf("[DBG]Run: No more items. Stopping main loop\n")
				break mainLoop
			}
			cmd.lastMsg = msg.DeepCopy()
			count++
		case <-ctx.Done():
			fmt.Printf("[DBG]Run: context cancelled. Stopping main loop\n")
			break mainLoop
		}
		// fmt.Printf("[DBG] len(outCh)=%d\n", len(outCh))
	}
	fmt.Printf("[DBG]Total entry processed: %d\n", count)
	cmd.Print(printer.PrinterModeJSON)
	fmt.Println("[DBG] Goodbye parse <file> match <whitelist>")
	cancel()

	return nil
}

func (cmd *MatchCmd) Print(mode printer.PrinterMode) error {
	var rcerror error

	rcerror = errortree.Add(rcerror, "matchcmd.Print", fmt.Errorf("printer mode %v not supported", mode))

	switch mode {
	case printer.PrinterModeJSON:
		rcerror = printJSON(cmd.lastMsg)
	case printer.PrinterModeTable:
	case printer.PrinterModeText:
	}

	return rcerror
}

func printJSON(msg isplunk.SplunkPipeMsg) error {
	type iplist struct {
		Len int
		IPs []string
	}
	type ips struct {
		Active  iplist
		Unknown iplist
	}
	type outputJSON struct {
		IPs ips
	}
	var rcerror error
	var jsonData outputJSON
	var activeIPs, unknownIPs iplist

	if activeips, ok := msg.Get(stage.ActiveIps).(map[string]bool); !ok {
		return errortree.Add(rcerror, "printOutputJSON", errors.New("data type mismatch"))
	} else {
		activeIPs = iplist{
			Len: len(activeips),
			IPs: make([]string, 0, len(activeips)),
		}
		for ip, _ := range activeips {
			activeIPs.IPs = append(activeIPs.IPs, ip)
		}
		// sort the slice by keys
		sort.Strings(activeIPs.IPs)
	}

	if unknownips, ok := msg.Get(stage.UnknonwIps).(map[string]bool); !ok {
		return errortree.Add(rcerror, "printOutputJSON", errors.New("data type mismatch"))
	} else {
		unknownIPs = iplist{
			Len: len(unknownips),
			IPs: make([]string, 0, len(unknownips)),
		}
		for ip, _ := range unknownips {
			unknownIPs.IPs = append(unknownIPs.IPs, ip)
		}
		// sort the slice by keys
		sort.Strings(unknownIPs.IPs)
	}

	jsonData = outputJSON{
		IPs: ips{
			Active:  activeIPs,
			Unknown: unknownIPs,
		},
	}
	// Convert structs to JSON.
	if j, err := json.Marshal(jsonData); err != nil {
		return errortree.Add(rcerror, "printOutputJSON", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}
