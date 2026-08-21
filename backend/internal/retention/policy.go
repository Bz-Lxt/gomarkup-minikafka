package retention

// SegmentInfo is a read-only snapshot used to decide which files can go.
type SegmentInfo struct {
	Base      int64
	Bytes     int64
	CreatedMS int64
	Active    bool
}

type Policy struct {
	MaxBytes int64
	MaxAgeMS int64
}

func (p Policy) Enabled() bool {
	return p.MaxBytes > 0 || p.MaxAgeMS > 0
}

// Select returns indices of non-active segments that may be deleted.
// Always keeps the newest non-active segment plus the active one,
// so consumers can still catch up from at least one sealed file.
func (p Policy) Select(nowMS int64, segs []SegmentInfo) []int {
	if !p.Enabled() || len(segs) < 3 {
		return nil
	}
	var total int64
	for _, s := range segs {
		total += s.Bytes
	}
	var drop []int
	for i, s := range segs {
		if s.Active || i >= len(segs)-2 {
			continue
		}
		overSize := p.MaxBytes > 0 && total > p.MaxBytes
		tooOld := p.MaxAgeMS > 0 && s.CreatedMS > 0 && nowMS-s.CreatedMS > p.MaxAgeMS
		if overSize || tooOld {
			drop = append(drop, i)
			total -= s.Bytes
		}
	}
	return drop
}
