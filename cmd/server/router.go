package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"issue-pm/config"
	"issue-pm/controller"
	"issue-pm/middleware"
	"issue-pm/repository"
	"issue-pm/service"
	"os"
	"path/filepath"
)

func newRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	r := repository.New(db)
	s := service.New(r)
	ctl := controller.New(s, cfg)
	g := gin.New()
	g.Use(middleware.RequestID(), middleware.Recovery(), middleware.Logger(), middleware.BodyLimit(1<<20), middleware.CORS())
	g.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	webDir := "web"
	if _, err := os.Stat(filepath.Join(webDir, "index.html")); err != nil {
		webDir = filepath.Join("..", "..", "web")
	}
	g.StaticFile("/", filepath.Join(webDir, "index.html"))
	g.Static("/static", filepath.Join(webDir, "static"))
	api := g.Group("/api/v1")
	api.POST("/auth/register", ctl.Register)
	api.POST("/auth/login", ctl.Login)
	secured := api.Group("")
	secured.Use(middleware.JWT(cfg))
	secured.GET("/projects", ctl.ListProjects)
	secured.POST("/projects", ctl.CreateProject)
	secured.GET("/projects/:projectId/members", ctl.ListMembers)
	secured.POST("/projects/:projectId/members", ctl.AddMember)
	secured.GET("/projects/:projectId/issues", ctl.ListIssues)
	secured.POST("/projects/:projectId/issues", ctl.CreateIssue)
	secured.GET("/issues/:id", ctl.GetIssue)
	secured.PATCH("/issues/:id", ctl.UpdateIssue)
	secured.DELETE("/issues/:id", ctl.DeleteIssue)
	secured.POST("/issues/:id/comments", ctl.AddComment)
	secured.GET("/issues/:id/comments", ctl.ListComments)
	secured.DELETE("/comments/:commentId", ctl.DeleteComment)
	return g
}
