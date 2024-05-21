package utils

import (
	"sync/atomic"
)

type AtomicBool struct {
	val int32
}

func (b *AtomicBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&b.val, i)
}

func (b *AtomicBool) Get() bool {
	return atomic.LoadInt32(&b.val) != 0
}

func (b *AtomicBool) CompareAndSwap(old, new bool) bool {
	var iOld int32 = 0
	if old {
		iOld = 1
	}
	var iNew int32 = 0
	if new {
		iNew = 1
	}
	return atomic.CompareAndSwapInt32(&b.val, iOld, iNew)
}
