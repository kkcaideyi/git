package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"issue-pm/model"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation failed")
)

type Repository interface {
	CreateUser(context.Context, *model.User) error
	FindUserByEmail(context.Context, string) (*model.User, error)
	FindUser(context.Context, uuid.UUID) (*model.User, error)
	CreateProject(context.Context, *model.Project) error
	AddProjectMember(context.Context, *model.ProjectMember) error
	IsProjectMember(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	ListProjectsForUser(context.Context, uuid.UUID, bool, int, int) ([]model.Project, int64, error)
	GetProject(context.Context, uuid.UUID) (*model.Project, error)
	ListProjectMembers(context.Context, uuid.UUID) ([]model.User, error)
	CreateIssue(context.Context, *model.Issue) error
	GetIssue(context.Context, uuid.UUID) (*model.Issue, error)
	UpdateIssue(context.Context, *model.Issue, map[string]any) error
	DeleteIssue(context.Context, *model.Issue) error
	ListIssues(context.Context, uuid.UUID, string, *uuid.UUID, string, int, int, string) ([]model.Issue, int64, error)
	CreateComment(context.Context, *model.Comment) error
	GetComment(context.Context, uuid.UUID) (*model.Comment, error)
	DeleteComment(context.Context, *model.Comment) error
	ListComments(context.Context, uuid.UUID) ([]model.Comment, error)
	Audit(context.Context, *model.AuditLog) error
}

type Service struct{ R Repository }

func New(r Repository) *Service { return &Service{R: r} }

func (s *Service) Register(ctx context.Context, email, name, password, role string) (*model.User, error) {
	if role != "admin" {
		role = "user"
	}
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{Email: strings.ToLower(strings.TrimSpace(email)), Name: strings.TrimSpace(name), PasswordHash: string(h), Role: role}
	if err := s.R.CreateUser(ctx, u); err != nil {
		return nil, ErrConflict
	}
	return u, nil
}
func (s *Service) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	u, err := s.R.FindUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, ErrForbidden
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, ErrForbidden
	}
	return u, nil
}
func (s *Service) CreateProject(ctx context.Context, actor uuid.UUID, p *model.Project) error {
	p.CreatedBy = actor
	if err := s.R.CreateProject(ctx, p); err != nil {
		return ErrConflict
	}
	if err := s.R.AddProjectMember(ctx, &model.ProjectMember{ProjectID: p.ID, UserID: actor, Role: "owner"}); err != nil {
		return err
	}
	return s.audit(ctx, actor, "project.created", "project", p.ID, nil)
}
func (s *Service) ListProjects(ctx context.Context, actor uuid.UUID, admin bool, limit, offset int) ([]model.Project, int64, error) {
	return s.R.ListProjectsForUser(ctx, actor, admin, limit, offset)
}
func (s *Service) CanAccessProject(ctx context.Context, actor, projectID uuid.UUID, admin bool) (bool, error) {
	if admin {
		return true, nil
	}
	p, err := s.R.GetProject(ctx, projectID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, ErrNotFound
	}
	if err != nil {
		return false, err
	}
	if p.CreatedBy == actor {
		return true, nil
	}
	return s.R.IsProjectMember(ctx, projectID, actor)
}
func (s *Service) AddMember(ctx context.Context, actor, projectID, userID uuid.UUID, role string, admin bool) error {
	ok, err := s.CanAccessProject(ctx, actor, projectID, admin)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	if !admin {
		project, projectErr := s.R.GetProject(ctx, projectID)
		if projectErr != nil || project.CreatedBy != actor {
			return ErrForbidden
		}
	}
	if _, err = s.R.FindUser(ctx, userID); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	if role != "member" && role != "owner" {
		role = "member"
	}
	if err = s.R.AddProjectMember(ctx, &model.ProjectMember{ProjectID: projectID, UserID: userID, Role: role}); err != nil {
		return err
	}
	return s.audit(ctx, actor, "project.member_added", "project", projectID, map[string]any{"user_id": userID.String(), "role": role})
}
func (s *Service) ListMembers(ctx context.Context, actor, projectID uuid.UUID, admin bool) ([]model.User, error) {
	ok, err := s.CanAccessProject(ctx, actor, projectID, admin)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return s.R.ListProjectMembers(ctx, projectID)
}
func (s *Service) IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	return s.R.IsProjectMember(ctx, projectID, userID)
}
func (s *Service) CreateIssue(ctx context.Context, actor uuid.UUID, admin bool, i *model.Issue) error {
	ok, err := s.CanAccessProject(ctx, actor, i.ProjectID, admin)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	if len(i.CustomFields) == 0 {
		i.CustomFields = datatypes.JSON([]byte(`{}`))
	}
	if len(i.Labels) == 0 {
		i.Labels = datatypes.JSON([]byte(`[]`))
	}
	i.CreatedBy = actor
	if err := s.R.CreateIssue(ctx, i); err != nil {
		return err
	}
	return s.audit(ctx, actor, "issue.created", "issue", i.ID, nil)
}
func (s *Service) GetIssue(ctx context.Context, id uuid.UUID) (*model.Issue, error) {
	i, err := s.R.GetIssue(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return i, err
}
func (s *Service) GetIssueForUser(ctx context.Context, actor uuid.UUID, admin bool, id uuid.UUID) (*model.Issue, error) {
	i, err := s.GetIssue(ctx, id)
	if err != nil {
		return nil, err
	}
	ok, err := s.CanAccessProject(ctx, actor, i.ProjectID, admin)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return i, nil
}
func (s *Service) UpdateIssue(ctx context.Context, actor uuid.UUID, admin bool, id uuid.UUID, fields map[string]any) (*model.Issue, error) {
	i, err := s.GetIssueForUser(ctx, actor, admin, id)
	if err != nil {
		return nil, err
	}
	if err = s.R.UpdateIssue(ctx, i, fields); err != nil {
		return nil, err
	}
	if err = s.audit(ctx, actor, "issue.updated", "issue", id, fields); err != nil {
		return nil, err
	}
	return s.GetIssue(ctx, id)
}
func (s *Service) DeleteIssue(ctx context.Context, actor uuid.UUID, admin bool, id uuid.UUID) error {
	i, err := s.GetIssueForUser(ctx, actor, admin, id)
	if err != nil {
		return err
	}
	if err = s.R.DeleteIssue(ctx, i); err != nil {
		return err
	}
	return s.audit(ctx, actor, "issue.deleted", "issue", id, nil)
}
func (s *Service) ListIssues(ctx context.Context, actor uuid.UUID, admin bool, pid uuid.UUID, status string, aid *uuid.UUID, label string, limit, offset int, sort string) ([]model.Issue, int64, error) {
	ok, err := s.CanAccessProject(ctx, actor, pid, admin)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, ErrForbidden
	}
	return s.R.ListIssues(ctx, pid, status, aid, label, limit, offset, sort)
}
func (s *Service) AddComment(ctx context.Context, actor uuid.UUID, admin bool, c *model.Comment) error {
	if _, err := s.GetIssueForUser(ctx, actor, admin, c.IssueID); err != nil {
		return err
	}
	c.AuthorID = actor
	if err := s.R.CreateComment(ctx, c); err != nil {
		return err
	}
	return s.audit(ctx, actor, "comment.created", "comment", c.ID, nil)
}
func (s *Service) Comments(ctx context.Context, actor uuid.UUID, admin bool, issueID uuid.UUID) ([]model.Comment, error) {
	if _, err := s.GetIssueForUser(ctx, actor, admin, issueID); err != nil {
		return nil, err
	}
	return s.R.ListComments(ctx, issueID)
}
func (s *Service) DeleteComment(ctx context.Context, actor uuid.UUID, admin bool, id uuid.UUID) error {
	c, err := s.R.GetComment(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if _, err = s.GetIssueForUser(ctx, actor, admin, c.IssueID); err != nil {
		return err
	}
	if !admin && c.AuthorID != actor {
		return ErrForbidden
	}
	if err = s.R.DeleteComment(ctx, c); err != nil {
		return err
	}
	return s.audit(ctx, actor, "comment.deleted", "comment", id, nil)
}
func (s *Service) audit(ctx context.Context, actor uuid.UUID, action, kind string, id uuid.UUID, metadata any) error {
	var payload datatypes.JSON
	if metadata != nil {
		b, _ := json.Marshal(metadata)
		payload = datatypes.JSON(b)
	}
	return s.R.Audit(ctx, &model.AuditLog{ActorID: actor, Action: action, EntityType: kind, EntityID: id, Metadata: payload})
}
