package fusefs

import (
	"container/list"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// CacheTracker manages SSD cache eviction using an LRU (Least Recently Used) policy.
// It tracks which file content hashes were most recently accessed,
// and evicts the least recently used ones once the cache exceeds MaxFiles entries.
type CacheTracker struct {
	MaxFiles    int                   		// Maximum number of files allowed in the cache
	recentFiles *list.List            		// Doubly-linked list of hashes (front = most recent)
	hashes      map[string]*list.Element 	// Map from hash to its position in the list
	lock        sync.Mutex            		// Protects access to the cache tracker state
}

// Global cache tracker instance shared across file reads.
// Set to allow up to 10 cached entries.
var cacheTracker = CacheTracker{
	MaxFiles:    10,
	recentFiles: list.New(),
	hashes:      make(map[string]*list.Element),
}

// Touch marks a file content hash as recently used.
// If the hash is new and the cache exceeds MaxFiles,
// the least recently used entry is evicted.
func (ct *CacheTracker) Touch(hash string) {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if entry, found := ct.hashes[hash]; found {
		// Move to front if already tracked
		ct.recentFiles.MoveToFront(entry)
		return
	}

	// Add to front if new
	newEntry := ct.recentFiles.PushFront(hash)
	ct.hashes[hash] = newEntry

	// Evict if over size limit
	if ct.recentFiles.Len() > ct.MaxFiles {
		ct.evictOldest()
	}
}

// evictOldest removes the least recently used file from SSDCache and metadataIndex.
// This function is only called when the number of cached entries exceeds MaxFiles.
func (ct *CacheTracker) evictOldest() {
	oldest := ct.recentFiles.Back()
	if oldest == nil {
		return
	}

	hash := oldest.Value.(string)
	ct.recentFiles.Remove(oldest)
	delete(ct.hashes, hash)

	cachePath := filepath.Join(SSDCache, hash)
	log.Printf("🗑️ Evicting %s from cache", cachePath)
	os.Remove(cachePath)

	metadataLock.Lock()
	for path, entry := range metadataIndex {
		if entry.Hash == hash {
			delete(metadataIndex, path)
			log.Printf("🧹 Removed %s from metadata index", path)
		}
	}
	metadataLock.Unlock()
	saveMetadataToFile()
}