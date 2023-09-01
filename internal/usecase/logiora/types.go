package logiora

type CLI struct {
	Parse ParseCmd `kong:"cmd, help:'Parses nginx logs'"`
}
