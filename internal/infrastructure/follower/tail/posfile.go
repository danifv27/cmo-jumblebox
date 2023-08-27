package tail

// PositionFile interface
type PositionFile interface {
	// Close closes this PositionFile
	Close() error
	// FileStat returns FileStat
	FileStat() *FileStat
	// Offset returns offset value
	Offset() int64
	// IncreaseOffset increases offset value
	IncreaseOffset(incr int) error
	// Set set fileStat and offset
	Set(fileStat *FileStat, offset int64) error
	// SetOffset set offset value
	SetOffset(offset int64) error
	// SetFileStat set fileStat
	SetFileStat(fileStat *FileStat) error
}
