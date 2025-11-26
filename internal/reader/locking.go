package reader

func (reader *ReaderImpl) rLock() {
	reader.lock.RLock()
}

func (reader *ReaderImpl) rUnlock() {
	reader.lock.RUnlock()
}

func (reader *ReaderImpl) rwLock() {
	reader.lock.Lock()
}

func (reader *ReaderImpl) rwUnlock() {
	reader.lock.Unlock()
}
