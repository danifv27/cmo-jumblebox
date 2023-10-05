package logiora

type ParseCmdFlags struct {
	Output string `kong:"help='Output format (json|excel|text)',enum='json,excel,text',default='json'"`
}
