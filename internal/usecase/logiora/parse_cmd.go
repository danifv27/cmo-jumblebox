package logiora

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	apipe "fry.org/qft/jumble/internal/application/pipeline"
	"fry.org/qft/jumble/internal/application/pipeline/stage"
	ifollower "fry.org/qft/jumble/internal/infrastructure/follower/file"
	ipipe "fry.org/qft/jumble/internal/infrastructure/pipeline"
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/davecgh/go-spew/spew"
	"github.com/speijnik/go-errortree"
)

type ParseCmd struct {
	Format string `kong:"help='Log format',default='$remote_addr - $remote_user [$time_local] \"$request\" $status $body_bytes_sent \"$http_referer\" \"$http_user_agent\" rt=$request_time uct=\"$upstream_connect_time\" uht=\"$upstream_header_time\" urt=\"$upstream_response_time\" x_request-id=\"$http_x_request_id\" reqid=\"$reqid\"'"`
	File   struct {
		File  string   `kong:"arg,help='File to parse'"`
		Match MatchCmd `kong:"cmd,help='Check if ip is whitelisted'"`
	} `kong:"arg"`
}

type MatchCmd struct {
	Whitelist struct {
		Whitelist string `kong:"arg,help='Whitelisted ranges'"`
	} `kong:"arg"`
}

func do(ctx context.Context, cancel context.CancelFunc, pipe apipe.Piper[isplunk.SplunkPipeMsg], entries chan string, errs chan error) (<-chan isplunk.SplunkPipeMsg, chan struct{}, error) {
	var rcerror, err error
	var outCh <-chan isplunk.SplunkPipeMsg

	quit := make(chan struct{})
	inCh := make(chan isplunk.SplunkPipeMsg)
	go func(ct context.Context) {
		for {
			select {
			case entry, hasMore := <-entries:
				if !hasMore {
					// fmt.Println("[DBG]No more entries")
					close(quit)
					return
				}
				expected := isplunk.NewSplunkMessage("input.entry", nil)
				expected.Add("entry", entry)
				inCh <- expected
			case failure, hasMore := <-errs:
				if !hasMore || failure != nil {
					// fmt.Println("[DBG]No more errors")
					close(quit)
					return
				}
			case <-ctx.Done():
				fmt.Printf("[DBG]context cancelled. Stopping source goroutine\n")
				close(quit)
				return
			}
		}
	}(ctx)

	// Run the pipeline
	if outCh, err = pipe.Do(ctx, inCh); err != nil {
		cancel()
		return outCh, quit, errortree.Add(rcerror, "matchcmd.do", err)
	}

	return outCh, quit, nil
}

func (cmd *MatchCmd) Run(cli *CLI) error {
	var rcerror, err error
	var ppln apipe.Piper[isplunk.SplunkPipeMsg]
	var quit chan struct{}
	var msg isplunk.SplunkPipeMsg
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
	regexParserStg := stage.NewRegexParse(cli.Tail.Format)
	st := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](regexParserStg.Do)
	ppln.Next(st)
	// Check unique ip
	ipStg := stage.NewIpSet(true, "remote_addr")
	ipset := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](ipStg.Do)
	ppln.Next(ipset)

	// Pipeline source
	entriesCh, errorsCh, err := follow.Lines(ctx)
	if err != nil {
		cancel()
		return errortree.Add(rcerror, "matchcmd.Run", err)
	}

	if outCh, quit, err = do(ctx, cancel, ppln, entriesCh, errorsCh); err != nil {
		cancel()
		return errortree.Add(rcerror, "matchcmd.Run", err)
	}

	// Drain pipeline
	count := 1
mainLoop:
	for {
		select {
		case msg = <-outCh:
			fmt.Printf("[DBG]entry processed: %d\n", count)
			count++
		case <-ctx.Done():
			// fmt.Printf("[DBG]context cancelled. Stopping main loop\n")
			cancel()
			break mainLoop
		case <-quit:
			cancel()
			break mainLoop
		}
	}
	spew.Dump(msg)
	fmt.Println("[DBG] Goodbye parse <file> match <whitelist>")
	cancel()

	return nil
}
