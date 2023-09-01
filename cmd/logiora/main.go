package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fry.org/qft/jumble/internal/usecase/logiora"
	"github.com/alecthomas/kong"
	"github.com/speijnik/go-errortree"
)

func getConfigPaths() ([]string, error) {
	var rcerror, err error
	var ex string
	var paths []string

	for i, s := range os.Args {
		if (s == "--config") && (len(os.Args) > i) {
			paths = append(paths, os.Args[i+1])
		}
	}

	bin := filepath.Base(os.Args[0])
	if ex, err = os.Executable(); err != nil {
		return paths, errortree.Add(rcerror, "getConfigPaths", err)
	}
	exPath := filepath.Dir(ex)
	exBin := filepath.Base(ex)
	paths = append(paths, fmt.Sprintf("/etc/%s.json", bin))
	paths = append(paths, fmt.Sprintf("~/.%s.json", bin))
	paths = append(paths, fmt.Sprintf("%s/.%s.json", exPath, exBin))

	return paths, nil
}

func main() {
	var err, rcerror error
	var configs []string

	cli := logiora.CLI{}
	if configs, err = getConfigPaths(); err != nil {
		panic(err)
	}
	bin := filepath.Base(os.Args[0])
	//config file has precedence over envars
	ctx := kong.Parse(&cli,
		kong.Bind(&rcerror),
		kong.Name(bin),
		kong.Description("Log Parser"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree: true,
		}),
		// kong.TypeMapper(reflect.TypeOf([]common.K8sResource{}), common.K8sResource{}),
		kong.Configuration(kong.JSON, configs...),
	)
	//Run should create the job flow that will be executed as a sequence
	if err = ctx.Run(&cli); err != nil {
		rcerror = errortree.Add(rcerror, "context", err)
		rcerror = errortree.Add(rcerror, "cmd", fmt.Errorf("%s", ctx.Command()))
		rcerror = errortree.Add(rcerror, "msg", fmt.Errorf("can not execute '%s' command", ctx.Command()))
		ctx.FatalIfErrorf(rcerror)
	}
}
