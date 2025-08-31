package post

import (
	"backend/item"
	"time"
)

type Post struct {
	Command    *item.Command `json:"command"`
	File       *item.File    `json:"file"`
	Author     time.Time     `json:"author"`
	CreatedAt  time.Time     `json:"createdOn"`
	UpdatedAt  string        `json:"lastUpdated"`
	Tags       []string      `json:"tags"`
	Comments   []Comment
	ForkCount  int `json:"numberOfForks"`
	CloneCount int `json:"cloneCount"`
}

type Comment struct {
	UserName string `json:"username"`
	Comment  string `json:"comment"`
}
