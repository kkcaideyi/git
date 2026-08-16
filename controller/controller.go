package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"issue-pm/config"
	"issue-pm/model"
	"issue-pm/pkg/auth"
	"issue-pm/pkg/response"
	"issue-pm/service"
)

type Controller struct {
	s   *service.Service
	cfg *config.Config
}

func New(s *service.Service, cfg *config.Config) *Controller { return &Controller{s: s, cfg: cfg} }
func userID(c *gin.Context) uuid.UUID                        { id, _ := uuid.Parse(c.GetString("user_id")); return id }
func isAdmin(c *gin.Context) bool                            { return c.GetString("role") == "admin" }
func page(c *gin.Context) (int, int, error) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, service.ErrValidation
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 || offset > 100000 {
		return 0, 0, service.ErrValidation
	}
	return limit, offset, nil
}
func decode(c *gin.Context, dst any) error {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return service.ErrValidation
	}
	return nil
}
func serviceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		response.Error(c, 400, response.CodeInvalid, "invalid request parameters")
	case errors.Is(err, service.ErrForbidden):
		response.Error(c, 403, response.CodeForbidden, "you do not have access to this resource")
	case errors.Is(err, service.ErrNotFound):
		response.Error(c, 404, response.CodeNotFound, "resource not found")
	case errors.Is(err, service.ErrConflict):
		response.Error(c, 409, response.CodeConflict, "resource already exists")
	default:
		response.Error(c, 500, response.CodeInternal, "internal server error")
	}
}
func parseParamID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	return id, err == nil
}
func validStatus(v string) bool {
	switch v {
	case "open", "in_progress", "resolved", "closed":
		return true
	}
	return false
}
func validPriority(v string) bool {
	switch v {
	case "low", "medium", "high", "urgent":
		return true
	}
	return false
}
func validateJSONShape(v datatypes.JSON, wantArray bool) bool {
	if len(v) == 0 || !json.Valid(v) {
		return false
	}
	var x any
	if json.Unmarshal(v, &x) != nil {
		return false
	}
	if wantArray {
		_, ok := x.([]any)
		return ok
	}
	_, ok := x.(map[string]any)
	return ok
}

func (ctl *Controller) Register(c *gin.Context) {
	var in struct{ Email, Name, Password string }
	if err := decode(c, &in); err != nil || mailAddressInvalid(in.Email) || len([]rune(strings.TrimSpace(in.Name))) < 2 || len(in.Password) < 8 {
		response.Error(c, 400, response.CodeInvalid, "valid email, name and password(min 8) are required")
		return
	}
	u, err := ctl.s.Register(c, in.Email, in.Name, in.Password, "user")
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Created(c, gin.H{"id": u.ID, "email": u.Email, "name": u.Name, "role": u.Role})
}
func mailAddressInvalid(v string) bool {
	a, err := mail.ParseAddress(strings.TrimSpace(v))
	return err != nil || a.Address != strings.TrimSpace(v)
}
func (ctl *Controller) Login(c *gin.Context) {
	var in struct{ Email, Password string }
	if err := decode(c, &in); err != nil || in.Email == "" || in.Password == "" {
		response.Error(c, 400, response.CodeInvalid, "email and password are required")
		return
	}
	u, err := ctl.s.Authenticate(c, in.Email, in.Password)
	if err != nil {
		response.Error(c, 401, response.CodeUnauthorized, "invalid credentials")
		return
	}
	token, err := auth.Generate(u.ID.String(), u.Role, ctl.cfg.JWTSecret, ctl.cfg.JWTExpireHours)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.OK(c, gin.H{"token": token, "user": gin.H{"id": u.ID, "email": u.Email, "name": u.Name, "role": u.Role}})
}
func (ctl *Controller) CreateProject(c *gin.Context) {
	var in struct{ Name, Key, Description string }
	if err := decode(c, &in); err != nil || len([]rune(strings.TrimSpace(in.Name))) < 2 || len(in.Key) < 2 || len(in.Key) > 20 {
		response.Error(c, 400, response.CodeInvalid, "name and key are required; key length must be 2-20")
		return
	}
	p := &model.Project{Name: strings.TrimSpace(in.Name), Key: strings.ToUpper(strings.TrimSpace(in.Key)), Description: strings.TrimSpace(in.Description)}
	if err := ctl.s.CreateProject(c, userID(c), p); err != nil {
		serviceError(c, err)
		return
	}
	response.Created(c, p)
}
func (ctl *Controller) ListProjects(c *gin.Context) {
	limit, offset, err := page(c)
	if err != nil {
		serviceError(c, err)
		return
	}
	rows, total, err := ctl.s.ListProjects(c, userID(c), isAdmin(c), limit, offset)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.OK(c, gin.H{"items": rows, "total": total, "limit": limit, "offset": offset})
}
func (ctl *Controller) ListMembers(c *gin.Context) {
	pid, ok := parseParamID(c, "projectId")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid project id")
		return
	}
	rows, err := ctl.s.ListMembers(c, userID(c), pid, isAdmin(c))
	if err != nil {
		serviceError(c, err)
		return
	}
	response.OK(c, rows)
}
func (ctl *Controller) AddMember(c *gin.Context) {
	pid, ok := parseParamID(c, "projectId")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid project id")
		return
	}
	var in struct {
		UserID uuid.UUID `json:"user_id"`
		Role   string    `json:"role"`
	}
	if err := decode(c, &in); err != nil || in.UserID == uuid.Nil {
		response.Error(c, 400, response.CodeInvalid, "user_id is required")
		return
	}
	if err := ctl.s.AddMember(c, userID(c), pid, in.UserID, in.Role, isAdmin(c)); err != nil {
		serviceError(c, err)
		return
	}
	response.Created(c, gin.H{"project_id": pid, "user_id": in.UserID})
}
func (ctl *Controller) CreateIssue(c *gin.Context) {
	pid, ok := parseParamID(c, "projectId")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid project id")
		return
	}
	var in struct {
		Title, Description, Status, Priority string
		AssigneeID                           *uuid.UUID     `json:"assignee_id"`
		Labels                               datatypes.JSON `json:"labels"`
		CustomFields                         datatypes.JSON `json:"custom_fields"`
	}
	if err := decode(c, &in); err != nil || len([]rune(strings.TrimSpace(in.Title))) < 1 || len([]rune(in.Title)) > 255 {
		response.Error(c, 400, response.CodeInvalid, "title is required and must be <=255 chars")
		return
	}
	if in.Status == "" {
		in.Status = "open"
	}
	if in.Priority == "" {
		in.Priority = "medium"
	}
	if !validStatus(in.Status) || !validPriority(in.Priority) || !validateJSONShape(in.Labels, true) && len(in.Labels) > 0 || !validateJSONShape(in.CustomFields, false) && len(in.CustomFields) > 0 {
		response.Error(c, 400, response.CodeInvalid, "invalid status, priority, labels or custom_fields")
		return
	}
	if in.AssigneeID != nil {
		member, err := ctl.s.IsMember(c, pid, *in.AssigneeID)
		if err != nil || !member {
			response.Error(c, 400, response.CodeInvalid, "assignee must be a project member")
			return
		}
	}
	i := &model.Issue{ProjectID: pid, Title: strings.TrimSpace(in.Title), Description: strings.TrimSpace(in.Description), Status: in.Status, Priority: in.Priority, AssigneeID: in.AssigneeID, Labels: in.Labels, CustomFields: in.CustomFields}
	if err := ctl.s.CreateIssue(c, userID(c), isAdmin(c), i); err != nil {
		serviceError(c, err)
		return
	}
	response.Created(c, i)
}
func (ctl *Controller) ListIssues(c *gin.Context) {
	pid, ok := parseParamID(c, "projectId")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid project id")
		return
	}
	var aid *uuid.UUID
	if v := c.Query("assignee_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(c, 400, response.CodeInvalid, "invalid assignee_id")
			return
		}
		aid = &id
	}
	status, sortBy, label := c.Query("status"), c.DefaultQuery("sort", "created_at"), c.Query("label")
	if status != "" && !validStatus(status) || sortBy != "created_at" && sortBy != "updated_at" && sortBy != "priority" || len([]rune(label)) > 80 {
		response.Error(c, 400, response.CodeInvalid, "invalid status, sort or label")
		return
	}
	limit, offset, err := page(c)
	if err != nil {
		serviceError(c, err)
		return
	}
	rows, total, err := ctl.s.ListIssues(c, userID(c), isAdmin(c), pid, status, aid, label, limit, offset, sortBy)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.OK(c, gin.H{"items": rows, "total": total, "limit": limit, "offset": offset})
}
func (ctl *Controller) GetIssue(c *gin.Context) {
	id, ok := parseParamID(c, "id")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid issue id")
		return
	}
	i, err := ctl.s.GetIssueForUser(c, userID(c), isAdmin(c), id)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.OK(c, i)
}
func (ctl *Controller) UpdateIssue(c *gin.Context) {
	id, ok := parseParamID(c, "id")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid issue id")
		return
	}
	var fields map[string]any
	if err := decode(c, &fields); err != nil {
		serviceError(c, err)
		return
	}
	if err := validateUpdate(fields); err != nil {
		serviceError(c, err)
		return
	}
	if value, present := fields["assignee_id"]; present && value != nil {
		assigneeText, ok := value.(string)
		assigneeID, parseErr := uuid.Parse(assigneeText)
		issue, accessErr := ctl.s.GetIssueForUser(c, userID(c), isAdmin(c), id)
		if !ok || parseErr != nil || accessErr != nil {
			response.Error(c, 400, response.CodeInvalid, "invalid assignee_id")
			return
		}
		member, memberErr := ctl.s.IsMember(c, issue.ProjectID, assigneeID)
		if memberErr != nil || !member {
			response.Error(c, 400, response.CodeInvalid, "assignee must be a project member")
			return
		}
	}
	i, err := ctl.s.UpdateIssue(c, userID(c), isAdmin(c), id, fields)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.OK(c, i)
}
func validateUpdate(fields map[string]any) error {
	allowed := map[string]bool{"title": true, "description": true, "status": true, "priority": true, "assignee_id": true, "labels": true, "custom_fields": true}
	if len(fields) == 0 {
		return service.ErrValidation
	}
	for k := range fields {
		if !allowed[k] {
			return service.ErrValidation
		}
	}
	if v, ok := fields["title"].(string); ok && (strings.TrimSpace(v) == "" || len([]rune(v)) > 255) {
		return service.ErrValidation
	}
	if v, ok := fields["status"].(string); ok && !validStatus(v) {
		return service.ErrValidation
	}
	if v, ok := fields["priority"].(string); ok && !validPriority(v) {
		return service.ErrValidation
	}
	if v, ok := fields["labels"]; ok {
		raw, _ := json.Marshal(v)
		if !validateJSONShape(datatypes.JSON(raw), true) {
			return service.ErrValidation
		}
	}
	if v, ok := fields["custom_fields"]; ok {
		raw, _ := json.Marshal(v)
		if !validateJSONShape(datatypes.JSON(raw), false) {
			return service.ErrValidation
		}
	}
	return nil
}
func (ctl *Controller) DeleteIssue(c *gin.Context) {
	id, ok := parseParamID(c, "id")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid issue id")
		return
	}
	if err := ctl.s.DeleteIssue(c, userID(c), isAdmin(c), id); err != nil {
		serviceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (ctl *Controller) AddComment(c *gin.Context) {
	id, ok := parseParamID(c, "id")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid issue id")
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if err := decode(c, &in); err != nil || len([]rune(strings.TrimSpace(in.Body))) < 1 || len([]rune(in.Body)) > 10000 {
		response.Error(c, 400, response.CodeInvalid, "body is required and must be <=10000 chars")
		return
	}
	cm := &model.Comment{IssueID: id, Body: strings.TrimSpace(in.Body)}
	if err := ctl.s.AddComment(c, userID(c), isAdmin(c), cm); err != nil {
		serviceError(c, err)
		return
	}
	response.Created(c, cm)
}
func (ctl *Controller) ListComments(c *gin.Context) {
	id, ok := parseParamID(c, "id")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid issue id")
		return
	}
	rows, err := ctl.s.Comments(c, userID(c), isAdmin(c), id)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.OK(c, rows)
}
func (ctl *Controller) DeleteComment(c *gin.Context) {
	id, ok := parseParamID(c, "commentId")
	if !ok {
		response.Error(c, 400, response.CodeInvalid, "invalid comment id")
		return
	}
	if err := ctl.s.DeleteComment(c, userID(c), isAdmin(c), id); err != nil {
		serviceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
