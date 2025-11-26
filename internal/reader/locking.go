package reader

import (
	"fmt"
	"runtime"
	"sync"
)

const LOCK_SECTION_SIZE = uint64(1024)

func createLocks() []sync.RWMutex {
	numLocks := runtime.NumCPU() * 4
	locks := make([]sync.RWMutex, numLocks)
	return locks
}

func (reader *ReaderImpl) getLocksForRange(startIndexInclusive uint64, endIndexInclusive uint64) []*sync.RWMutex {
	if endIndexInclusive < startIndexInclusive {
		panic(fmt.Sprintf("Can't lock negative size range %d-%d", startIndexInclusive, endIndexInclusive))
	}

	lockIndex := (startIndexInclusive / LOCK_SECTION_SIZE) % uint64(len(reader.locks))
	lastLockIndex := (endIndexInclusive / LOCK_SECTION_SIZE) % uint64(len(reader.locks))

	var locksToUse []*sync.RWMutex
	for {
		locksToUse = append(locksToUse, &reader.locks[lockIndex])
		if lockIndex == lastLockIndex {
			break
		}
		lockIndex = (lockIndex + 1) % uint64(len(reader.locks))
	}

	return locksToUse
}

func (reader *ReaderImpl) rLock(startIndexInclusive uint64, endIndexInclusive uint64) {
	for _, lock := range reader.getLocksForRange(startIndexInclusive, endIndexInclusive) {
		lock.RLock()
	}
}

func (reader *ReaderImpl) rUnlock(startIndexInclusive uint64, endIndexInclusive uint64) {
	for _, lock := range reader.getLocksForRange(startIndexInclusive, endIndexInclusive) {
		lock.RUnlock()
	}
}

func (reader *ReaderImpl) rwLock(startIndexInclusive uint64, endIndexInclusive uint64) {
	for _, lock := range reader.getLocksForRange(startIndexInclusive, endIndexInclusive) {
		lock.Lock()
	}
}

func (reader *ReaderImpl) rwUnlock(startIndexInclusive uint64, endIndexInclusive uint64) {
	for _, lock := range reader.getLocksForRange(startIndexInclusive, endIndexInclusive) {
		lock.Unlock()
	}
}
