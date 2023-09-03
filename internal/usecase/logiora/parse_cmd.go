package logiora

import "fmt"

// Parse parts of an nginx log entry
type ParseCmd struct{}
type ParseCmd struct {
	File string `kong:"arg,type=existingfile,required,help='File to parse'"`
}

func (cmd *ParseCmd) Run(cli *CLI, rcerror *error) error {

	fmt.Printf("[DBg]%sHello World\n")
	fmt.Printf("[DBG]cli: %+v\n", *cli)
	fmt.Printf("[DBG]rcerror: %+v\n", *rcerror)

	return nil
}
