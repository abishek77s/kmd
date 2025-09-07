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

	_, err := DB.Exec(r.Context(),
		`INSERT INTO commands (id, data)
         VALUES ($1, $2)
         ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data`,
		cmd.ID, cmd)
	if err != nil {
		http.Error(w, "DB insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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

	var data []byte

	err := DB.QueryRow(r.Context(),
		`SELECT data FROM commands WHERE id=$1`, id).Scan(&data)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	var cmd item.Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		http.Error(w, "Invalid stored data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cmd)
}

func GetFileHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var data []byte

	err := DB.QueryRow(r.Context(),
		`SELECT data FROM files WHERE id=$1`, id).Scan(&data)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	var file item.File
	if err := json.Unmarshal(data, &file); err != nil {
		http.Error(w, "Invalid stored data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

func StoreFileHandler(w http.ResponseWriter, r *http.Request) {
	var file item.File
	if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if file.ID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	_, err := DB.Exec(r.Context(),
		`INSERT INTO files (id, data)
         VALUES ($1, $2)
         ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data`,
		file.ID, file)
	if err != nil {
		http.Error(w, "DB insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Stored command successfully",
		"id":      file.ID,
	})
}
