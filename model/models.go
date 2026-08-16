package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Name         string         `gorm:"size:120;not null" json:"name"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Role         string         `gorm:"size:20;not null;default:user;index" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Project struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"size:120;not null" json:"name"`
	Key         string         `gorm:"size:20;uniqueIndex;not null" json:"key"`
	Description string         `json:"description"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;index" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type ProjectMember struct {
	ProjectID uuid.UUID `gorm:"type:uuid;primaryKey" json:"project_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Role      string    `gorm:"size:20;not null;default:member" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
type Issue struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_issue_project_status" json:"project_id"`
	Title        string         `gorm:"size:255;not null" json:"title"`
	Description  string         `json:"description"`
	Status       string         `gorm:"size:30;not null;default:open;index:idx_issue_project_status" json:"status"`
	Priority     string         `gorm:"size:20;not null;default:medium" json:"priority"`
	AssigneeID   *uuid.UUID     `gorm:"type:uuid;index" json:"assignee_id"`
	Labels       datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"labels"`
	CustomFields datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"custom_fields"`
	CreatedBy    uuid.UUID      `gorm:"type:uuid;not null;index" json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Comment struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	IssueID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"issue_id"`
	AuthorID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"author_id"`
	Body      string         `gorm:"not null" json:"body"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type AuditLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ActorID    uuid.UUID      `gorm:"type:uuid;index"`
	Action     string         `gorm:"size:80;not null;index"`
	EntityType string         `gorm:"size:40;not null"`
	EntityID   uuid.UUID      `gorm:"type:uuid;index"`
	Metadata   datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt  time.Time
}

func ensureUUID(id *uuid.UUID) {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}
func (m *User) BeforeCreate(_ *gorm.DB) error     { ensureUUID(&m.ID); return nil }
func (m *Project) BeforeCreate(_ *gorm.DB) error  { ensureUUID(&m.ID); return nil }
func (m *Issue) BeforeCreate(_ *gorm.DB) error    { ensureUUID(&m.ID); return nil }
func (m *Comment) BeforeCreate(_ *gorm.DB) error  { ensureUUID(&m.ID); return nil }
func (m *AuditLog) BeforeCreate(_ *gorm.DB) error { ensureUUID(&m.ID); return nil }
