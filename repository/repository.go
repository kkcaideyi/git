package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"issue-pm/model"
)

type Repo struct{ DB *gorm.DB }

func New(db *gorm.DB) *Repo { return &Repo{DB: db} }
func (r *Repo) CreateUser(ctx context.Context, u *model.User) error {
	return r.DB.WithContext(ctx).Create(u).Error
}
func (r *Repo) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := r.DB.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *Repo) FindUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	if err := r.DB.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *Repo) CreateProject(ctx context.Context, p *model.Project) error {
	return r.DB.WithContext(ctx).Create(p).Error
}
func (r *Repo) AddProjectMember(ctx context.Context, m *model.ProjectMember) error {
	return r.DB.WithContext(ctx).Where("project_id = ? AND user_id = ?", m.ProjectID, m.UserID).FirstOrCreate(m).Error
}
func (r *Repo) IsProjectMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&model.ProjectMember{}).Where("project_id = ? AND user_id = ?", projectID, userID).Count(&n).Error
	return n > 0, err
}
func (r *Repo) ListProjectsForUser(ctx context.Context, userID uuid.UUID, admin bool, limit, offset int) ([]model.Project, int64, error) {
	var rows []model.Project
	var total int64
	q := r.DB.WithContext(ctx).Model(&model.Project{})
	if !admin {
		q = q.Where("created_by = ? OR id IN (SELECT project_id FROM project_members WHERE user_id = ?)", userID, userID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
func (r *Repo) GetProject(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var p model.Project
	if err := r.DB.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *Repo) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]model.User, error) {
	var users []model.User
	err := r.DB.WithContext(ctx).Table("users u").Select("u.id, u.email, u.name, u.role, u.created_at, u.updated_at").Joins("JOIN project_members pm ON pm.user_id = u.id").Where("pm.project_id = ?", projectID).Order("u.name ASC").Scan(&users).Error
	return users, err
}
func (r *Repo) CreateIssue(ctx context.Context, i *model.Issue) error {
	return r.DB.WithContext(ctx).Create(i).Error
}
func (r *Repo) GetIssue(ctx context.Context, id uuid.UUID) (*model.Issue, error) {
	var i model.Issue
	if err := r.DB.WithContext(ctx).First(&i, id).Error; err != nil {
		return nil, err
	}
	return &i, nil
}
func (r *Repo) UpdateIssue(ctx context.Context, i *model.Issue, fields map[string]any) error {
	return r.DB.WithContext(ctx).Model(i).Updates(fields).Error
}
func (r *Repo) DeleteIssue(ctx context.Context, i *model.Issue) error {
	return r.DB.WithContext(ctx).Delete(i).Error
}
func (r *Repo) ListIssues(ctx context.Context, projectID uuid.UUID, status string, assignee *uuid.UUID, label string, limit, offset int, sort string) ([]model.Issue, int64, error) {
	var rows []model.Issue
	var total int64
	q := r.DB.WithContext(ctx).Model(&model.Issue{}).Where("project_id = ?", projectID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if assignee != nil {
		q = q.Where("assignee_id = ?", *assignee)
	}
	if label != "" {
		payload, _ := json.Marshal([]string{label})
		q = q.Where("labels @> ?::jsonb", string(payload))
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "created_at"
	switch sort {
	case "updated_at":
		order = "updated_at"
	case "priority":
		order = "priority"
	}
	if err := q.Order(order + " DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
func (r *Repo) CreateComment(ctx context.Context, c *model.Comment) error {
	return r.DB.WithContext(ctx).Create(c).Error
}
func (r *Repo) GetComment(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	var c model.Comment
	if err := r.DB.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
func (r *Repo) DeleteComment(ctx context.Context, c *model.Comment) error {
	return r.DB.WithContext(ctx).Delete(c).Error
}
func (r *Repo) ListComments(ctx context.Context, id uuid.UUID) ([]model.Comment, error) {
	var cs []model.Comment
	err := r.DB.WithContext(ctx).Where("issue_id = ?", id).Order("created_at ASC, id ASC").Find(&cs).Error
	return cs, err
}
func (r *Repo) Audit(ctx context.Context, a *model.AuditLog) error {
	return r.DB.WithContext(ctx).Create(a).Error
}
