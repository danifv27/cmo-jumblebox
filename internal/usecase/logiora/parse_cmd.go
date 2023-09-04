package logiora

import (
	"context"
	"fmt"
	"os"
	"time"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"
	apipe "fry.org/qft/jumble/internal/application/pipeline"
	iTail "fry.org/qft/jumble/internal/infrastructure/follower/tail"
	ipipe "fry.org/qft/jumble/internal/infrastructure/pipeline"
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"

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
	// var fStat *iTail.FileStat
	// var positionFile iTail.PositionFile

	if ppln, err = ipipe.Parse[isplunk.SplunkPipeMsg]("pipeline:splunk"); err != nil {
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
	//The source of the pipeline are the lines from the log file
	if f, err = os.OpenFile(cli.Parse.File, os.O_RDONLY|os.O_SYNC, 0600); err != nil {
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}
	// if fStat, err = iTail.Stat(f); err != nil {
	// 	return errortree.Add(rcerror, "parsecmd.Run", err)
	// }
	//TODO: compute position file from temp dir
	// if positionFile, err = iTail.OpenPositionFile("./tailPosition.file"); err != nil {
	// 	return errortree.Add(rcerror, "parsecmd.Run", err)
	// }
	configs := []aconfigurable.ConfigurablerFn{
		// iTail.WithPositionFile(positionFile),
		iTail.WithDetectRotateDelay(time.Second),
		iTail.WithWatchRotateInterval(10 * time.Millisecond),
		// iTail.WithFollowRotate(true),
	}
	follow, e := iTail.NewTailFollower(f.Name(), configs...)
	if e != nil {
		return errortree.Add(rcerror, "parsecmd.Run", err)
	}

	fmt.Printf("[DBG]Hello World\n")
	fmt.Printf("[DBG]cli: %+v\n", *cli)
	fmt.Printf("[DBG]ppln: %+v\n", ppln)

	// Create a new context
	ctx := context.Background()

	logEntriesCh, logErrorCh := follow.Lines(ctx)

waitForCancellation:
	for {
		select {
		case logEntry := <-logEntriesCh:
			fmt.Printf("[DBG]entry: %s\n", logEntry)
		case logError := <-logErrorCh:
			fmt.Printf("[DBG]error: %v\n", logError)
			return errortree.Add(rcerror, "parseCmd.Run", err)
		case <-ctx.Done():
			fmt.Printf("[DBG]context cancelled\n")
			break waitForCancellation
		}
	}
	fmt.Printf("[DBG]Goodbye World\n")

	return nil
}

// CreateFile creates a file in the temp dir
// func (d *TempDir) CreateFile(name string) (*os.File, *FileStat) {
// 	f, err := os.OpenFile(filepath.Join(d.Path, name), os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0600)
// 	if err != nil {
// 		panic(err)
// 	}
// 	s, err := Stat(f)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return f, s
// }
