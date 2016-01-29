package lg

import "sync"

var (
	memlogmu sync.RWMutex
	memlog   []string
)

func logToMemory(line []byte) {
	if len(line) < 1 {
		return
	}
	const (
		maxLength = 50000
		cutN      = 1000
	)

	str := string(line)
	str = str[0 : len(str)-1]
	memlogmu.Lock()
	memlog = append(memlog, str)
	if len(memlog) > maxLength {
		memlog = memlog[cutN:maxLength]
	}
	memlogmu.Unlock()
}

// Memlog returns the in memory log file
func Memlog() []string {
	memlogmu.RLock()
	lines := make([]string, len(memlog))
	copy(lines, memlog)
	memlogmu.RUnlock()
	return lines
}
