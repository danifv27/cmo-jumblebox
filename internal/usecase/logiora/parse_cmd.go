package logiora

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	apipe "fry.org/qft/jumble/internal/application/pipeline"
	"fry.org/qft/jumble/internal/application/pipeline/stage"
	iTail "fry.org/qft/jumble/internal/infrastructure/follower/un000/tailor"
	ipipe "fry.org/qft/jumble/internal/infrastructure/pipeline"
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"

	"github.com/davecgh/go-spew/spew"
	"github.com/speijnik/go-errortree"
)

// Parse parts of an nginx log entry
type ParseCmd struct {
	Format string `kong:"help='Log format',default='$remote_addr - $remote_user [$time_local] \"$request\" $status $body_bytes_sent \"$http_referer\" \"$http_user_agent\" rt=$request_time uct=\"$upstream_connect_time\" uht=\"$upstream_header_time\" urt=\"$upstream_response_time\" x_request-id=\"$http_x_request_id\" reqid=\"$reqid\"'"`
	File   string `kong:"arg,type=existingfile,required,help='File to parse'"`
}

func (cmd *ParseCmd) Run(cli *CLI) error {
	var rcerror, err error
	var ppln apipe.Piper[isplunk.SplunkPipeMsg]
	var f *os.File
	var outCh <-chan isplunk.SplunkPipeMsg

	if ppln, err = ipipe.Parse[isplunk.SplunkPipeMsg]("pipeline:splunk"); err != nil {
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
	regexParserStg := stage.NewRegexParse(cli.Parse.Format)

	//The source of the pipeline are the lines from the log file
	if f, err = os.OpenFile(cli.Parse.File, os.O_RDONLY|os.O_SYNC, 0600); err != nil {
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
	follow := iTail.NewFollower(f.Name())
	// ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// Pipeline source
	logEntriesCh, logErrorCh, err := follow.Lines(ctx)
	if err != nil {
		cancel()
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
	// Prepare the pipeline
	st := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](regexParserStg.Do)
	ppln.Next(st)

	ipStg := stage.NewIpSet(true, "remote_addr")
	ipset := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](ipStg.Do)
	ppln.Next(ipset)
	inCh := make(chan isplunk.SplunkPipeMsg)

	go func(ct context.Context) {
		for {
			select {
			case logEntry := <-logEntriesCh:
				expected := isplunk.NewSplunkMessage("tail.entry", nil)
				expected.Add("entry", logEntry)
				inCh <- expected
			case logError := <-logErrorCh:
				fmt.Printf("[DBG]error: %v\n", logError)
				cancel()
			case <-ct.Done():
				// fmt.Printf("[DBG]context cancelled. Stopping source goroutine\n")
				return
			}
		}
	}(ctx)

	// Run the pipeline
	if outCh, err = ppln.Do(ctx, inCh); err != nil {
		cancel()
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
	count := 1
	var msg isplunk.SplunkPipeMsg
mainLoop:
	// Drain pipelines
	for {
		select {
		case msg = <-outCh:
			//
			fmt.Printf("[DBG]ip processed: %d\n", count)
			count++
		case <-ctx.Done():
			// fmt.Printf("[DBG]context cancelled. Stopping main loop\n")
			break mainLoop
		}
	}

	spew.Dump(msg)
	fmt.Printf("[DBG]Goodbye World\n")
	cancel()

	return nil
}
