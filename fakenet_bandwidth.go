package raft

// Bandwidth 带宽的结构体定义,bit数,每ms时间,unit换算单位(默认1000)
type Bandwidth struct {
	bits float64
	ms   float64
	unit float64
}

func NewBandwidth(bits float64, ms float64, u ...float64) Bandwidth {
	// 检查换算单位
	unit := float64(1000)
	if len(u) != 0 {
		unit = u[0]
	}
	return Bandwidth{
		bits: bits,
		ms:   ms,
		unit: unit,
	}
}

func NewBandwidthWithMbps(bw float64, u ...float64) Bandwidth {
	// 检查换算单位
	unit := float64(1000)
	if len(u) != 0 {
		unit = u[0]
	}
	return Bandwidth{
		bits: bw * unit * unit, // 1Mb = 1 * unit * unit bit
		ms:   1000,             // per second
		unit: unit,             // 1000
	}
}
func NewBW(bw float64, u ...float64) Bandwidth {
	// 检查换算单位
	unit := float64(1000)
	if len(u) != 0 {
		unit = u[0]
	}
	return Bandwidth{
		bits: bw * unit * unit, // 1Mb = 1 * unit * unit bit
		ms:   1000,             // per second
		unit: unit,             // 1000
	}
}

func NewBandwidthWithBitsAndMs(bits float64, ms float64, u ...float64) Bandwidth {
	// 检查换算单位
	unit := float64(1000)
	if len(u) != 0 {
		unit = u[0]
	}
	return Bandwidth{
		bits: bits,
		ms:   ms,
		unit: unit,
	}
}

// bits per ms
func (b *Bandwidth) bpms() float64 {
	return b.bits / b.ms
}

// K bits per ms
func (b *Bandwidth) Kbpms() float64 {
	return b.bpms() / b.unit
}

// M bits per ms
func (b *Bandwidth) Mbpms() float64 {
	return b.Kbpms() / b.unit
}

// bits per ms
func (b *Bandwidth) bps() float64 {
	return b.bpms() * 1000
}

// K bits per ms
func (b *Bandwidth) Kbps() float64 {
	return b.bps() / b.unit
}

// M bits per ms
func (b *Bandwidth) Mbps() float64 {
	return b.Kbps() / b.unit
}

// Bytes per ms
func (b *Bandwidth) Bpms() float64 {
	return b.bpms() / 8
}

// K Bytes per ms
func (b *Bandwidth) KBpms() float64 {
	return b.Bpms() / b.unit
}

// M Bytes per ms
func (b *Bandwidth) MBpms() float64 {
	return b.KBpms() / b.unit
}

// Bytes per s
func (b *Bandwidth) Bps() float64 {
	return b.Bpms() * 1000
}

// K Bytes per s
func (b *Bandwidth) KBps() float64 {
	return b.Bps() / b.unit
}

// M Bytes per s
func (b *Bandwidth) MBps() float64 {
	return b.KBps() / b.unit
}

//// Bandwidth 表示每秒传输速率
//type Bandwidth struct {
//	toMb         float64
//	toMbPerMs    float64
//	toKb         float64
//	toKbPerMs    float64
//	toBits       float64
//	toBitsPerMs  float64
//	toMB         float64
//	toMBPerMs    float64
//	toKB         float64
//	toKBPerMs    float64
//	toBytes      float64
//	toBytesPerMs float64
//}
//
//func NewBandwidth(n uint64,) Bandwidth { // 1Mbps = 1000 Kbps = 1000000 bps
//	return Bandwidth{
//		toMb:         n,
//		toKb:         n * 1000,
//		toBits:       n * 1000 * 1000,
//		toMB:         n / 8,
//		toKB:         n * 1000 / 8,
//		toBytes:      n * 1000 * 1000 / 8,
//		toBytesPerMs: n * 1000 / 8,
//	}
//}
//
//func NewBandwidthFromBitsAndMs(bits float64, ms float64) Bandwidth {
//	return NewBandwidth(bits / 1000 / 1000)
//}
