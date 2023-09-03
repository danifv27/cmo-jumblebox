package logiora

type CLI struct {
	Parse ParseCmd `kong:"cmd,help='Parse log files'"`
}
