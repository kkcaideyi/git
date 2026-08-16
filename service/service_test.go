package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"issue-pm/model"
)

type fakeRepo struct {
	Repository
	users    map[uuid.UUID]*model.User
	projects map[uuid.UUID]*model.Project
	members  map[[2]uuid.UUID]bool
	issues   map[uuid.UUID]*model.Issue
	comments map[uuid.UUID]*model.Comment
	audits   []*model.AuditLog
}

func newFake() *fakeRepo {
	return &fakeRepo{users: map[uuid.UUID]*model.User{}, projects: map[uuid.UUID]*model.Project{}, members: map[[2]uuid.UUID]bool{}, issues: map[uuid.UUID]*model.Issue{}, comments: map[uuid.UUID]*model.Comment{}}
}
func (f *fakeRepo) CreateUser(_ context.Context, u *model.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	f.users[u.ID] = u
	return nil
}
func (f *fakeRepo) FindUser(_ context.Context, id uuid.UUID) (*model.User, error) {
	return f.users[id], nil
}
func (f *fakeRepo) CreateProject(_ context.Context, p *model.Project) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	f.projects[p.ID] = p
	return nil
}
func (f *fakeRepo) AddProjectMember(_ context.Context, m *model.ProjectMember) error {
	f.members[[2]uuid.UUID{m.ProjectID, m.UserID}] = true
	return nil
}
func (f *fakeRepo) IsProjectMember(_ context.Context, p, u uuid.UUID) (bool, error) {
	return f.members[[2]uuid.UUID{p, u}], nil
}
func (f *fakeRepo) GetProject(_ context.Context, id uuid.UUID) (*model.Project, error) {
	return f.projects[id], nil
}
func (f *fakeRepo) ListProjectsForUser(context.Context, uuid.UUID, bool, int, int) ([]model.Project, int64, error) {
	return nil, 0, nil
}
func (f *fakeRepo) CreateIssue(_ context.Context, i *model.Issue) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	f.issues[i.ID] = i
	return nil
}
func (f *fakeRepo) GetIssue(_ context.Context, id uuid.UUID) (*model.Issue, error) {
	return f.issues[id], nil
}
func (f *fakeRepo) UpdateIssue(_ context.Context, i *model.Issue, fields map[string]any) error {
	for k, v := range fields {
		switch k {
		case "status":
			i.Status = v.(string)
		}
	}
	return nil
}
func (f *fakeRepo) DeleteIssue(_ context.Context, i *model.Issue) error {
	delete(f.issues, i.ID)
	return nil
}
func (f *fakeRepo) CreateComment(_ context.Context, c *model.Comment) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	f.comments[c.ID] = c
	return nil
}
func (f *fakeRepo) GetComment(_ context.Context, id uuid.UUID) (*model.Comment, error) {
	return f.comments[id], nil
}
func (f *fakeRepo) DeleteComment(_ context.Context, c *model.Comment) error {
	delete(f.comments, c.ID)
	return nil
}
func (f *fakeRepo) Audit(_ context.Context, a *model.AuditLog) error {
	f.audits = append(f.audits, a)
	return nil
}

func TestProjectIsolationAndAudit(t *testing.T) {
	ctx := context.Background()
	r := newFake()
	s := New(r)
	owner, other := uuid.New(), uuid.New()
	p := &model.Project{Name: "Private", Key: "PRV"}
	if err := s.CreateProject(ctx, owner, p); err != nil {
		t.Fatal(err)
	}
	if ok, _ := s.CanAccessProject(ctx, owner, p.ID, false); !ok {
		t.Fatal("owner should access")
	}
	if ok, _ := s.CanAccessProject(ctx, other, p.ID, false); ok {
		t.Fatal("non-member should be denied")
	}
	if len(r.audits) != 1 || r.audits[0].ActorID != owner {
		t.Fatalf("expected project audit actor, got %#v", r.audits)
	}
}
func TestIssueUpdateRequiresProjectAccess(t *testing.T) {
	ctx := context.Background()
	r := newFake()
	s := New(r)
	owner, other := uuid.New(), uuid.New()
	p := &model.Project{Name: "P", Key: "P1"}
	_ = s.CreateProject(ctx, owner, p)
	i := &model.Issue{ProjectID: p.ID, Title: "bug"}
	_ = s.CreateIssue(ctx, owner, false, i)
	if _, err := s.UpdateIssue(ctx, other, false, i.ID, map[string]any{"status": "resolved"}); err != ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if _, err := s.UpdateIssue(ctx, owner, false, i.ID, map[string]any{"status": "resolved"}); err != nil {
		t.Fatal(err)
	}
}
func TestCommentDeleteOwnerOrAdmin(t *testing.T) {
	ctx := context.Background()
	r := newFake()
	s := New(r)
	owner, author, other := uuid.New(), uuid.New(), uuid.New()
	p := &model.Project{Name: "P", Key: "P2"}
	_ = s.CreateProject(ctx, owner, p)
	_ = s.AddMember(ctx, owner, p.ID, author, "member", false)
	i := &model.Issue{ProjectID: p.ID, Title: "bug"}
	_ = s.CreateIssue(ctx, owner, false, i)
	c := &model.Comment{IssueID: i.ID, Body: "note"}
	if err := s.AddComment(ctx, author, false, c); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteComment(ctx, other, false, c.ID); err != ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if err := s.DeleteComment(ctx, author, false, c.ID); err != nil {
		t.Fatal(err)
	}
}
