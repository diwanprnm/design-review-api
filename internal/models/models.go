package models

import "time"

// Vocabulary mirrors CONTEXT.md. DTOs match docs/api.md shapes.

// Designer dipetakan GORM ke tabel "designers" (konvensi GORM:
// struct Designer -> tabel plural snake_case, tanpa config tambahan).
// Skema tabel tetap berasal dari migrations/0001_init.sql — tag gorm
// di sini hanya petunjuk pemetaan, bukan pembuat skema.
type Designer struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	Name         string    `gorm:"not null" json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ShareToken     *string   `json:"share_token,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Design struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	LatestVersion *VersionBrief  `json:"latest_version,omitempty"`
}

type VersionBrief struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}

type Version struct {
	ID        string    `json:"id"`
	DesignID  string    `json:"design_id"`
	Number    int       `json:"number"`
	ImageURL  string    `json:"image_url"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	CreatedAt time.Time `json:"created_at"`
}

// Pin carries relative % coords (Q13) and a denormalized root comment for lists.
type Pin struct {
	ID          string       `json:"id"`
	VersionID   string       `json:"version_id"`
	X           float64      `json:"x"`
	Y           float64      `json:"y"`
	Status      string       `json:"status"` // open|resolved|reopened
	ReplyCount  int          `json:"reply_count"`
	RootComment *Comment     `json:"root_comment,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Comment: parent_id nil = root; else single-level reply to root.
type Comment struct {
	ID          string    `json:"id"`
	PinID       string    `json:"pin_id"`
	ParentID    *string   `json:"parent_id"`
	Body        string    `json:"body"`
	AuthorType  string    `json:"author_type"` // client|designer
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
