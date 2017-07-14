package cli

import "github.com/dustin/go-humanize"

type MaxBandWidthType uint64

func (mb *MaxBandWidthType) String() string {
	return humanize.Bytes(uint64(*mb))
}

func (mb *MaxBandWidthType) Type() string {
	return "MaxBandWidth"
}

func (mb *MaxBandWidthType) Set(v string) error {
	if v == "" {
		*mb = MaxBandWidthType(0)
		return nil
	}
	if bw, err := humanize.ParseBytes(v); err != nil {
		return err
	} else {
		*mb = MaxBandWidthType(bw)
	}
	return nil
}
