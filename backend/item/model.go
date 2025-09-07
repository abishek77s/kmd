package item

type CacheItem struct {
	Command    *Command
	File       *File
	Expiration int64
}

type Command struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Commands []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"commands"`
}

type File struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}
