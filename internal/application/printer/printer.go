package printer

type PrinterMode int

const (
	PrinterModeNone  PrinterMode = iota //0
	PrinterModeJSON                     //1
	PrinterModeText                     //2
	PrinterModeTable                    //3
)

type Printer interface {
	Print(mode PrinterMode) error
}
