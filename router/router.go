package router

import (
	"tu-xun/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.Use(middleware.Error)
	r.Use(middleware.GinLogger(), middleware.GinRecovery(true))

	// 静态文件服务 —— 上传的图片可通过 URL 直接访问
	r.Static("/uploads", "./uploads")

	apiRouter := r.Group("/api")
	{

		// --- 用户认证 ---
		authRouter := apiRouter.Group("/auth")
		{
			authRouter.POST("/register", ctr.Auth.Register)
			authRouter.POST("/login", ctr.Auth.Login)
			authRouter.DELETE("/logout", ctr.Auth.Logout)
			authRouter.GET("/me", ctr.Auth.Me)
		}

		// --- 图片（图寻题目） ---
		photoRouter := apiRouter.Group("/photos")
		{
			photoRouter.GET("", ctr.Photo.List)       // 公共浏览
			photoRouter.GET("/:id", ctr.Photo.Detail) // 图片详情
			photoRouter.POST("", ctr.Photo.Upload)    // 上传投稿（需登录）

			// 答题
			photoRouter.POST("/:id/attempts", ctr.Attempt.Submit)       // 提交答案（需登录）
			photoRouter.GET("/:id/my-attempts", ctr.Attempt.MyAttempts) // 我的答题（需登录）

			// 故事
			photoRouter.POST("/:id/stories", ctr.Story.Create)     // 发布故事（需登录）
			photoRouter.GET("/:id/stories", ctr.Story.ListByPhoto) // 故事列表
		}

		// --- 我的奖品 ---
		apiRouter.GET("/users/me/prizes", ctr.Prize.MyPrizes)

		// --- 管理员接口 ---
		adminRouter := apiRouter.Group("/admin")
		adminRouter.Use(middleware.CheckRole(1))
		{
			adminRouter.GET("/photos/pending", ctr.Admin.PendingPhotos)
			adminRouter.PUT("/photos/:id/review", ctr.Admin.ReviewPhoto)
			adminRouter.GET("/attempts/pending", ctr.Admin.PendingAttempts)
			adminRouter.PUT("/attempts/:id/review", ctr.Admin.ReviewAttempt)
			adminRouter.PUT("/prizes/:id/claim", ctr.Admin.ClaimPrize)
		}
	}
}
