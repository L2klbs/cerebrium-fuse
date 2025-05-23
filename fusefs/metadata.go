package fusefs

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// CacheEntry tracks metadata for a cached file.
// - Hash: the content hash of the file in SSDCache
// - Timestamp: the last modification time of the source file in NFSRoot
type CacheEntry struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
}

// metadataIndex maps relative NFS paths to their cached hash and timestamp.
// Used to validate cache freshness across file reads.
var (
	metadataIndex = make(map[string]CacheEntry)
	metadataLock  sync.Mutex
)

// saveMetadataToFile writes the current metadataIndex to a JSON file.
// This is called after cache changes to persist metadata across restarts.
func saveMetadataToFile() error {
	metadataLock.Lock()
	defer metadataLock.Unlock()

	file, err := os.Create("cache_metadata.json")
	if err != nil {
		return err
	}
	defer file.Close()

	e := json.NewEncoder(file)
	e.SetIndent("", "  ")
	return e.Encode(metadataIndex)
}

// loadMetadataFromFile loads the metadataIndex from disk at startup.
// If the file doesn't exist, it silently skips loading.
func loadMetadataFromFile() error {
	metadataLock.Lock()
	defer metadataLock.Unlock()

	file, err := os.Open("cache_metadata.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(&metadataIndex)
}