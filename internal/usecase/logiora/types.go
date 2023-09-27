package logiora

type CLI struct {
	Tail  TailCmd  `kong:"cmd,help='Tail log files'"`
	Parse ParseCmd `kong:"cmd,help='Parse log files'"`
}
