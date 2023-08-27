package tail

// InMemory creates a inMemory PositionFile
func InMemory(fileStat *FileStat, offset int64) PositionFile {

	return &inMemory{entry{FileStat: fileStat, Offset: offset}}
}

type inMemory struct {
	entry
}

func (pf *inMemory) Close() error {

	return nil
}

func (pf *inMemory) FileStat() *FileStat {

	return pf.entry.FileStat
}

func (pf *inMemory) Offset() int64 {

	return pf.entry.Offset
}

func (pf *inMemory) IncreaseOffset(incr int) error {

	return pf.Set(pf.FileStat(), pf.Offset()+int64(incr))
}

func (pf *inMemory) Set(fileStat *FileStat, offset int64) error {

	pf.entry.FileStat = fileStat
	pf.entry.Offset = offset
	return nil
}

func (pf *inMemory) SetOffset(offset int64) error {

	return pf.Set(pf.FileStat(), offset)
}

func (pf *inMemory) SetFileStat(fileStat *FileStat) error {

	return pf.Set(fileStat, pf.Offset())
}
