package tail

import (
	"encoding/gob"
	"os"
)

type entry struct {
	FileStat *FileStat
	Offset   int64
}

// OpenPositionFile opens named PositionFile
func OpenPositionFile(name string) (PositionFile, error) {

	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0600)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	var ent entry
	if fi.Size() == 0 {
		return &positionFile{f: f, entry: ent}, nil
	}
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&ent); err != nil {
		return nil, err
	}

	return &positionFile{f: f, entry: ent}, nil
}

type positionFile struct {
	f *os.File
	entry
}

func (pf *positionFile) Close() error {

	return pf.f.Close()
}

func (pf *positionFile) FileStat() *FileStat {

	return pf.entry.FileStat
}

func (pf *positionFile) Offset() int64 {

	return pf.entry.Offset
}

func (pf *positionFile) IncreaseOffset(incr int) error {

	return pf.Set(pf.FileStat(), pf.Offset()+int64(incr))
}

func (pf *positionFile) Set(fileStat *FileStat, offset int64) error {

	pf.entry.FileStat = fileStat
	pf.entry.Offset = offset

	if _, err := pf.f.Seek(0, 0); err != nil {
		return err
	}
	enc := gob.NewEncoder(pf.f)
	if err := enc.Encode(&pf.entry); err != nil {
		return err
	}

	return nil
}

func (pf *positionFile) SetOffset(offset int64) error {

	return pf.Set(pf.FileStat(), offset)
}

func (pf *positionFile) SetFileStat(fileStat *FileStat) error {

	return pf.Set(fileStat, pf.Offset())
}
