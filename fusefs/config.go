package fusefs

// NFSRoot is the root directory for the simulated slow storage.
// All file lookups and reads originate from here unless cached.
var NFSRoot = "./nfs"

// SSDCache is the root directory for the simulated fast cache.
// You can use this to cache frequently accessed files from NFSRoot.
var SSDCache = "./ssd"