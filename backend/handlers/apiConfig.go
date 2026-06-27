package handlers

import "sync/atomic"

type ApiConfig struct {
	FileserverHits atomic.Int32
}
