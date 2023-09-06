package logiora

import (
	"context"
	"fmt"
	"os"

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
	File string `kong:"arg,type=existingfile,required,help='File to parse'"`
}

func (cmd *ParseCmd) Run(cli *CLI) error {
	var rcerror, err error
	var ppln apipe.Piper[isplunk.SplunkPipeMsg]
	var f *os.File
	var outCh <-chan isplunk.SplunkPipeMsg

	// var fStat *iTail.FileStat
	// var positionFile iTail.PositionFile

	if ppln, err = ipipe.Parse[isplunk.SplunkPipeMsg]("pipeline:splunk"); err != nil {
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
	//TODO: parametrise parsing regular expression
	regexParser := stage.NewRegexParse("$remote_addr - $remote_user [$time_local] \"$request\" $status $body_bytes_sent \"$http_referer\" \"$http_user_agent\" rt=$request_time uct=\"$upstream_connect_time\" uht=\"$upstream_header_time\" urt=\"$upstream_response_time\" x_request-id=\"$http_x_request_id\" reqid=\"$reqid\"")

	st := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](regexParser.Parse)
	inCh := make(chan isplunk.SplunkPipeMsg)
	ppln.Next(st)
	//The source of the pipeline are the lines from the log file
	if f, err = os.OpenFile(cli.Parse.File, os.O_RDONLY|os.O_SYNC, 0600); err != nil {
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
	follow := iTail.NewFollower(f.Name())

	// fmt.Printf("[DBG]Hello World\n")
	// fmt.Printf("[DBG]cli: %+v\n", *cli)
	// fmt.Printf("[DBG]ppln: %+v\n", ppln)

	// Create a new context
	ctx, cancel := context.WithCancel(context.Background())
	if outCh, err = ppln.Do(ctx, inCh); err != nil {
		cancel()
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}

	// Pipeline source
	logEntriesCh, logErrorCh, err := follow.Lines(ctx)
	if err != nil {
		cancel()
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
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
				fmt.Printf("[DBG]context cancelled\n")
				return
			}
		}
	}(ctx)

mainLoop:
	// Drain pipeline
	for {
		select {
		case msg := <-outCh:
			spew.Dump("[DBG]msg:", msg)
			// fmt.Printf("[DBG]msg: %+v\n", msg)
		case <-ctx.Done():
			fmt.Printf("[DBG]context cancelled\n")
			break mainLoop
		}
	}

	fmt.Printf("[DBG]Goodbye World\n")
	cancel()

	return nil
}
