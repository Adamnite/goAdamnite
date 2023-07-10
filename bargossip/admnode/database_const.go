package admnode

import "time"

const (
	dbVersionKey         = "version"
	dbNodeKeyPrefix      = "gn:"
	dbNodePongPrefix     = "pong"
	dbNodePingPrefix     = "ping"
	dbNodeFindFailPrefix = "findfail"
)

const (
	dbVersion = 0
)

const (
	dbCleanupDuration = 1 * time.Hour
	dbNodeExpireTime  = 72 * time.Hour
)
