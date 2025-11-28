package reader

import (
	"fmt"
	"runtime"
	"sync"
)

const LOCK_SECTION_SIZE = 1024

func createLocks() []sync.RWMutex {
	numLocks := runtime.NumCPU() * 4
	locks := make([]sync.RWMutex, numLocks)
	return locks
}

func (reader *ReaderImpl) getLocksForRange(startIndexInclusive int, endIndexInclusive int) []*sync.RWMutex {
	if endIndexInclusive < startIndexInclusive {
		panic(fmt.Sprintf("Can't lock negative size range %d-%d", startIndexInclusive, endIndexInclusive))
	}

	firstLockSectionIndex := (startIndexInclusive / LOCK_SECTION_SIZE)
	lastLockSectionIndex := (endIndexInclusive / LOCK_SECTION_SIZE)

	// FIXME: Does make() here impact benchmark numbers? Otherwise maybe skip
	// it?
	locksToUse := make([]*sync.RWMutex, 0, len(reader.lineLocks))

	lockSectionsCount := lastLockSectionIndex - firstLockSectionIndex + 1
	if lockSectionsCount >= len(reader.lineLocks) {
		// Range is larger than number of locks, must use all locks
		for i := 0; i < len(reader.lineLocks); i++ {
			locksToUse = append(locksToUse, &reader.lineLocks[i])
		}
		return locksToUse
	}

	firstLockIndex := firstLockSectionIndex % len(reader.lineLocks)
	lastLockIndex := lastLockSectionIndex % len(reader.lineLocks)

	// Normal case, just list the needed locks
	lockIndex := firstLockIndex
	for {
		locksToUse = append(locksToUse, &reader.lineLocks[lockIndex])
		if lockIndex == lastLockIndex {
			break
		}
		lockIndex = (lockIndex + 1) % len(reader.lineLocks)
	}

	return locksToUse
}

func (reader *ReaderImpl) rLock(startIndexInclusive int, endIndexInclusive int) {
	for _, lock := range reader.getLocksForRange(startIndexInclusive, endIndexInclusive) {
		lock.RLock()
	}
}

func (reader *ReaderImpl) rUnlock(startIndexInclusive int, endIndexInclusive int) {
	for _, lock := range reader.getLocksForRange(startIndexInclusive, endIndexInclusive) {
		lock.RUnlock()
	}
}

func (reader *ReaderImpl) rwLock(startIndexInclusive int, endIndexInclusive int) {
	for _, lock := range reader.getLocksForRange(startIndexInclusive, endIndexInclusive) {
		lock.Lock()
	}
}

func (reader *ReaderImpl) rwUnlock(startIndexInclusive int, endIndexInclusive int) {
	for _, lock := range reader.getLocksForRange(startIndexInclusive, endIndexInclusive) {
		lock.Unlock()
	}
}
