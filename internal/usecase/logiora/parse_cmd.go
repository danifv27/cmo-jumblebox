package logiora

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	alogger "fry.org/qft/jumble/internal/application/logger"
	apipe "fry.org/qft/jumble/internal/application/pipeline"
	"fry.org/qft/jumble/internal/application/pipeline/stage"
	"fry.org/qft/jumble/internal/application/printer"
	ifollower "fry.org/qft/jumble/internal/infrastructure/follower/file"
	ipipe "fry.org/qft/jumble/internal/infrastructure/pipeline"
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/speijnik/go-errortree"
	"github.com/tidwall/pretty"
	"github.com/xuri/excelize/v2"
)

type ParseCmd struct {
	Format string        `kong:"help='Log format to be parsed',default='$http_x_original_forwarded_for $remote_addr - $remote_user [$time_local] - $request $status'"`
	Flags  ParseCmdFlags `kong:"embed"`
	File   struct {
		File  string   `kong:"arg,help='File to parse'"`
		Match MatchCmd `kong:"cmd,help='Check if ip is whitelisted'"`
	} `kong:"arg"`
	lg alogger.Logger
}

type MatchCmd struct {
	Flags     MatchCmdFlags `kong:"embed"`
	Whitelist struct {
		Whitelist string `kong:"arg,help='Ranges to whitelist (cidr notation)'"`
	} `kong:"arg"`
	lastMsg isplunk.SplunkPipeMsg
}

type MatchCmdFlags struct {
	Cwd string `kong:"type='existingdir',default='./',help='Specifies the current working directory'"`
}

func do(lg alogger.Logger, ctx context.Context, cancel context.CancelFunc, pipe apipe.Piper[isplunk.SplunkPipeMsg], entries chan string, errs chan error) (<-chan isplunk.SplunkPipeMsg, error) {
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
					lg.Debug("No more entries\n")
					return
				}
				expected := isplunk.NewSplunkMessage("input.entry", nil)
				expected.Add("entry", entry)
				inCh <- expected
			case failure, hasMore := <-errs:
				if !hasMore || failure != nil {
					lg.Debug("No more errors\n")
					return
				}
			case <-ctx.Done():
				lg.Debug("Context cancelled. Stopping source goroutine\n")
				return
			}
		}
	}(ctx)

	// Run the pipeline
	if outCh, err = pipe.Do(ctx, inCh); err != nil {
		lg.Debugf("pipe.Do err: %s\n", err)
		cancel()
		return outCh, errortree.Add(rcerror, "matchcmd.do", err)
	}

	return outCh, nil
}

func (cmd *ParseCmd) Run(cli *CLI, lg alogger.Logger) error {
	var rcerror, err error
	var ppln apipe.Piper[isplunk.SplunkPipeMsg]
	var outCh <-chan isplunk.SplunkPipeMsg

	if cli.Debug {
		lg.SetLevel(alogger.LoggerLevelDebug)
	}
	cmd.lg = lg

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

	if outCh, err = do(lg, ctx, cancel, ppln, entriesCh, errorsCh); err != nil {
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
				lg.Debug("Run: No more items. Stopping main loop\n")
				break mainLoop
			}
			cmd.File.Match.lastMsg = msg.DeepCopy()
			count++
		case <-ctx.Done():
			lg.Debug("Run: Context cancelled. Stopping main loop\n")
			break mainLoop
		}
	}
	lg.Debugf("Total entry processed: %d\n", count)
	m := cli.Parse.Flags.Output
	err = nil
	switch {
	case m == "json":
		err = cmd.Print(printer.PrinterModeJSON)
	case m == "excel":
		err = cmd.Print(printer.PrinterModeExcel)
	case m == "text":
		err = cmd.Print(printer.PrinterModeText)
	}

	lg.Debug("Goodbye parse <file> match <whitelist>\n")
	cancel()

	return err
}

func (cmd *ParseCmd) Print(mode printer.PrinterMode) error {
	var rcerror error

	rcerror = errortree.Add(rcerror, "matchcmd.Print", fmt.Errorf("printer mode %v not supported", mode))

	switch mode {
	case printer.PrinterModeJSON:
		rcerror = printJSON(cmd.File.Match.lastMsg)
	case printer.PrinterModeExcel:
		rcerror = printExcel(cmd.File.Match.lastMsg, cmd.lg, cmd.File.Match.Flags.Cwd)
	case printer.PrinterModeTable:
	case printer.PrinterModeText:
	}

	return rcerror
}

func printExcel(msg isplunk.SplunkPipeMsg, lg alogger.Logger, cwd string) error {
	var rcerror, err error
	var addr string
	var headerStyle, titleStyle, subheaderStyle int

	t := time.Now()
	f := excelize.NewFile()
	defer func() {
		f.Close()
	}()
	sheetName := "Sheet1"
	// f.SetSheetName(sheetName, "PDM-Flex")

	//header values
	header := map[int][]interface{}{
		1: {"Log Analysis"},
		3: {"Logged IPs", "", "Allowed CIDR", ""},
		4: {"Active", "Not Whitelisted", "Used", "Not Used"},
	}
	// custom rows height
	height := map[int]float64{
		1: 36, 2: 10, 3: 28, 4: 28,
	}

	//Set header value
	for r, row := range header {
		if addr, err = excelize.JoinCellName("B", r); err != nil {
			return errortree.Add(rcerror, "printExcel", err)
		}
		if err = f.SetSheetRow(sheetName, addr, &row); err != nil {
			return errortree.Add(rcerror, "printExcel", err)
		}
	}
	// set custom row height
	for r, ht := range height {
		if err = f.SetRowHeight(sheetName, r, ht); err != nil {
			return errortree.Add(rcerror, "printExcel", err)
		}
	}
	// set custom column width
	if err = f.SetColWidth(sheetName, "B", "E", 24); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}

	// Title
	// merge cell for the 'Analisys'
	if err = f.MergeCell(sheetName, "B1", "E1"); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}
	// define font style for the 'Analisys'
	if titleStyle, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "1f7f3b", Bold: true, Size: 22, Family: "Arial"},
	}); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}
	// set font style for the 'Analisys'
	if err = f.SetCellStyle(sheetName, "B1", "E1", titleStyle); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}

	// Table headers
	// merge cell for the 'headers'
	if err = f.MergeCell(sheetName, "B3", "C3"); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}
	if err = f.MergeCell(sheetName, "D3", "E3"); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}
	// define style for the headers
	if headerStyle, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "1f7f3b", Bold: true, Family: "Arial"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E6F4EA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center"},
		Border:    []excelize.Border{{Type: "top", Style: 2, Color: "1f7f3b"}},
	}); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}
	// set style for the headers
	if err = f.SetCellStyle(sheetName, "B3", "E3", headerStyle); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}
	// define style for the headers
	if subheaderStyle, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "1f7f3b", Bold: true, Family: "Arial"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E6F4EA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "top", Style: 2, Color: "1f7f3b"},
			{Type: "bottom", Style: 2, Color: "1f7f3b"},
		},
	}); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}
	// set style for the subheaders
	if err = f.SetCellStyle(sheetName, "B4", "E4", subheaderStyle); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}

	if activeips, ok := msg.Get(stage.AllowedIpsStageKey).(map[string]bool); ok {
		idx := 4
		for ip := range activeips {
			idx++
			if addr, err = excelize.JoinCellName("B", idx); err != nil {
				return errortree.Add(rcerror, "printExcel", err)
			}
			if err = f.SetCellStr(sheetName, addr, ip); err != nil {
				return errortree.Add(rcerror, "printExcel", err)
			}
		}
	} else {
		return errortree.Add(rcerror, "printExcel", errors.New("data type mismatch"))
	}
	if unknownips, ok := msg.Get(stage.UnknonwIpsStageKey).(map[string]bool); ok {
		idx := 4
		for ip := range unknownips {
			idx++
			if addr, err = excelize.JoinCellName("C", idx); err != nil {
				return errortree.Add(rcerror, "printExcel", err)
			}
			if err = f.SetCellStr(sheetName, addr, ip); err != nil {
				return errortree.Add(rcerror, "printExcel", err)
			}
		}
	} else {
		return errortree.Add(rcerror, "printExcel", errors.New("data type mismatch"))
	}
	if activecidrs, ok := msg.Get(stage.ActiveSubnetsStageKey).(map[string]bool); ok {
		idx := 4
		for ip := range activecidrs {
			idx++
			if addr, err = excelize.JoinCellName("D", idx); err != nil {
				return errortree.Add(rcerror, "printExcel", err)
			}
			if err = f.SetCellStr(sheetName, addr, ip); err != nil {
				return errortree.Add(rcerror, "printExcel", err)
			}
		}
	} else {
		return errortree.Add(rcerror, "printExcel", errors.New("data type mismatch"))
	}
	if inactivecidrs, ok := msg.Get(stage.ActiveSubnetsStageKey).(map[string]bool); ok {
		idx := 4
		for ip := range inactivecidrs {
			idx++
			if addr, err = excelize.JoinCellName("E", idx); err != nil {
				return errortree.Add(rcerror, "printExcel", err)
			}
			if err = f.SetCellStr(sheetName, addr, ip); err != nil {
				return errortree.Add(rcerror, "printExcel", err)
			}
		}
	} else {
		return errortree.Add(rcerror, "printExcel", errors.New("data type mismatch"))
	}
	// Save spreadsheet by the given path.
	filename := filepath.Join(cwd, fmt.Sprintf("parsedIps-%s.xlsx", t.Format("2006-01-02T15:04:05Z")))
	if err := f.SaveAs(filename); err != nil {
		return errortree.Add(rcerror, "printExcel", err)
	}
	lg.Debugf("%s created succesfully", filename)

	return nil
}

func printJSON(msg isplunk.SplunkPipeMsg) error {
	type iplist struct {
		Len int
		IPs []string
	}
	type ips struct {
		Allowed iplist
		Unknown iplist
	}
	type cidrs struct {
		Active   iplist
		Inactive iplist
	}
	type outputJSON struct {
		IPs         ips
		Whitelisted cidrs
	}
	var rcerror error
	var jsonData outputJSON
	var activeIPs, unknownIPs iplist
	var activeCIDRs, inactiveCIDRs iplist

	if activeips, ok := msg.Get(stage.AllowedIpsStageKey).(map[string]bool); ok {
		activeIPs = iplist{
			Len: len(activeips),
			IPs: make([]string, 0, len(activeips)),
		}
		for ip := range activeips {
			activeIPs.IPs = append(activeIPs.IPs, ip)
		}
		// sort the slice by keys
		sort.Strings(activeIPs.IPs)
	} else {
		return errortree.Add(rcerror, "printJSON", errors.New("data type mismatch"))
	}
	if unknownips, ok := msg.Get(stage.UnknonwIpsStageKey).(map[string]bool); ok {
		unknownIPs = iplist{
			Len: len(unknownips),
			IPs: make([]string, 0, len(unknownips)),
		}
		for ip := range unknownips {
			unknownIPs.IPs = append(unknownIPs.IPs, ip)
		}
		// sort the slice by keys
		sort.Strings(unknownIPs.IPs)
	} else {
		return errortree.Add(rcerror, "printJSON", errors.New("data type mismatch"))
	}

	if activecidrs, ok := msg.Get(stage.ActiveSubnetsStageKey).(map[string]bool); ok {
		activeCIDRs = iplist{
			Len: len(activecidrs),
			IPs: make([]string, 0, len(activecidrs)),
		}
		for ip := range activecidrs {
			activeCIDRs.IPs = append(activeCIDRs.IPs, ip)
		}
		// sort the slice by keys
		sort.Strings(activeCIDRs.IPs)
	} else {
		return errortree.Add(rcerror, "printJSON", errors.New("data type mismatch"))
	}

	if inactivecidrs, ok := msg.Get(stage.InactiveSubnetsStageKey).(map[string]bool); ok {
		inactiveCIDRs = iplist{
			Len: len(inactivecidrs),
			IPs: make([]string, 0, len(inactivecidrs)),
		}
		for ip := range inactivecidrs {
			inactiveCIDRs.IPs = append(inactiveCIDRs.IPs, ip)
		}
		// sort the slice by keys
		sort.Strings(inactiveCIDRs.IPs)
	} else {
		return errortree.Add(rcerror, "printJSON", errors.New("data type mismatch"))
	}

	jsonData = outputJSON{
		IPs: ips{
			Allowed: activeIPs,
			Unknown: unknownIPs,
		},
		Whitelisted: cidrs{
			Active:   activeCIDRs,
			Inactive: inactiveCIDRs,
		},
	}
	// Convert structs to JSON.
	if j, err := json.Marshal(jsonData); err != nil {
		return errortree.Add(rcerror, "printJSON", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}
