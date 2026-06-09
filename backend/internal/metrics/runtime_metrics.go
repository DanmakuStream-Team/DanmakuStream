package metrics

import "sync/atomic"

var activeVideoConnections int64

func BeginVideoConnection() {
	atomic.AddInt64(&activeVideoConnections, 1)
}

func EndVideoConnection() {
	atomic.AddInt64(&activeVideoConnections, -1)
}

func ActiveVideoConnections() int64 {
	value := atomic.LoadInt64(&activeVideoConnections)
	if value < 0 {
		return 0
	}
	return value
}
