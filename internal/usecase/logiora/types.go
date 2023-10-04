package logiora

type CLI struct {
	Debug bool     `kong:"short='D',help='Enable debug mode'"`
	Tail  TailCmd  `kong:"cmd,help='Tail log files'"`
	Parse ParseCmd `kong:"cmd,help='Parse log files'"`
}
