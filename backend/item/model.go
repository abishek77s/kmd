package item

type CacheItem struct {
	Command    *Command
	File       *File
	Expiration int64
}

type Command struct {
	ID   string `json:"id"`
	Cmds []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"cmds"`
}

type File struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}
