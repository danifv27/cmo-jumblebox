package logiora

import "fmt"

type ParseCmd struct {
	Format string `kong:"help='Log format',default='$http_x_original_forwarded_for $remote_addr - $remote_user [$time_local] - $request $status'"`
	File   struct {
		// File   string `kong:"arg,type=existingfile,required,help='File to parse'"`
		File  string   `kong:"arg,help='File to parse',required,type=existingfile"` // <-- NOTE: identical name to enclosing struct field.
		Match MatchCmd `kong:"cmd,help='Check if ip is whitelisted'"`
	} `kong:"arg"`
}

type MatchCmd struct {
	Whitelist struct {
		Whitelist string `kong:"arg,help='Whitelisted ranges',required,type=existingfile"`
	} `kong:"arg"`
}

func (cmd *ParseCmd) Run(cli *CLI) error {

	fmt.Println("[DBG]Parse cmd: Hola caracola")

	return nil
}

func (cmd *MatchCmd) Run(cli *CLI) error {

	fmt.Println("[DBG]Match cmd: Hola caracola")

	return nil
}
