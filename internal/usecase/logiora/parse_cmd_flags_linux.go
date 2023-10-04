package logiora

type ParseCmdFlags struct {
	Output string `kong:"help='Output format (json|text)',enum='json',default='json,text'"`
}
