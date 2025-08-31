package api

import (
	"backend/item"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	cacheStore = make(map[string]item.CacheItem)
	mu         sync.RWMutex
	ttl        = 10 * time.Minute
)

// POST /store
func StoreCommandsHandler(w http.ResponseWriter, r *http.Request) {
	var cmd item.Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if cmd.ID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	cacheStore[cmd.ID] = item.CacheItem{
		Command:    &cmd,
		Expiration: time.Now().Add(ttl).UnixNano(),
	}
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Stored command successfully",
		"id":      cmd.ID,
	})

}

func GetCommandsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	mu.RLock()
	item, found := cacheStore[id]
	mu.RUnlock()

	if !found || time.Now().UnixNano() > item.Expiration {
		http.Error(w, "Not found or expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item.Command)
}

func GetFileHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	mu.RLock()
	item, found := cacheStore[id]
	mu.RUnlock()

	if !found || time.Now().UnixNano() > item.Expiration {
		http.Error(w, "Not found or expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item.File)
}

func StoreFileHandler(w http.ResponseWriter, r *http.Request) {
	var code item.File
	if err := json.NewDecoder(r.Body).Decode(&code); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if code.ID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	cacheStore[code.ID] = item.CacheItem{
		File:       &code,
		Expiration: time.Now().Add(ttl).UnixNano(),
	}
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Stored command successfully",
		"id":      code.ID,
	})
}
