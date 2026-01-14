package router

import (
	"cloud_notes/internal/config"
	"cloud_notes/internal/handler"
	"cloud_notes/internal/middleware"
	"cloud_notes/internal/repository"
	"cloud_notes/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	// 用户模块
	api := r.Group("/api")
	{
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)
	}

	// 登录
	auth := api.Group("")
	auth.Use(middleware.JWTAuth())

	// 笔记模块
	noteRepo := repository.NewNoteRepository(config.DB)
	notebookRepo := repository.NewNotebookRepository(config.DB)
	noteService := service.NewNoteService(noteRepo, notebookRepo)
	noteHandler := handler.NewNoteHandler(noteService)

	auth.POST("/notes", noteHandler.Create)
	auth.GET("/notes", noteHandler.List)
	auth.PUT("/notes/:id", noteHandler.Update)
	auth.DELETE("/notes/:id", noteHandler.Delete)

	// 笔记本模块
	notebookService := service.NewNotebookService(config.DB, notebookRepo, noteRepo)
	notebookHandler := handler.NewNotebookHandler(notebookService)

	auth.POST("/notebooks", notebookHandler.Create)
	auth.GET("/notebooks", notebookHandler.List)
	auth.PUT("/notebooks/:id", notebookHandler.Update)
	auth.DELETE("/notebooks/:id", notebookHandler.Delete)

	// 标签模块
	tagRepo := repository.NewTagRepository(config.DB)
	tagService := service.NewTagService(tagRepo, noteRepo)
	tagHandler := handler.NewTagHandler(tagService)

	auth.POST("/tags", tagHandler.Create)
	auth.GET("/tags", tagHandler.List)
	auth.POST("/notes/:id/tags", tagHandler.BindNoteTags)
	auth.GET("/tags/:id/notes", tagHandler.GetNotesByTag)
	auth.DELETE("/tags/:id", tagHandler.Delete)
}
