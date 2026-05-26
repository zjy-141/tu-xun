package router

import (
	"tu-xun/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.Use(middleware.Error)
	r.Use(middleware.GinLogger(), middleware.GinRecovery(true))

	apiRouter := r.Group("/api")
	{

		// --- 用户认证 ---
		authRouter := apiRouter.Group("/auth")
		{
			authRouter.POST("/register", ctr.Auth.Register)
			authRouter.POST("/login", ctr.Auth.Login)
			authRouter.DELETE("/logout", ctr.Auth.Logout)
			authRouter.Use(middleware.CheckRole(0)) // 下面功能需要登录
			authRouter.GET("/me", ctr.Auth.Me)
			authRouter.PUT("/password", ctr.Auth.ChangePassword)
			authRouter.PUT("/profile", ctr.Auth.UpdateProfile)
			authRouter.PUT("/description", ctr.Auth.UpdateDescription)
			authRouter.POST("/avatar", ctr.Auth.UploadAvatar)
		}

		// --- 图片（图寻题目） ---
		photoRouter := apiRouter.Group("/photos")
		{
			photoRouter.GET("", ctr.Photo.List)                       // 公共浏览
			photoRouter.GET("/:id", ctr.Photo.Detail)                 // 图片详情
			photoRouter.GET("/:id/image", ctr.Photo.GetImageStream)   // 图片展示（流式）
			photoRouter.GET("/:id/download", ctr.Photo.Download)      // 图片下载
			photoRouter.GET("/:id/comments", ctr.Photo.PhotoComments) // 图片评论列表
			photoRouter.GET("/:id/attempts", ctr.Photo.PhotoAttempts) // 图片答题列表
			photoRouter.Use(middleware.CheckRole(0))
			photoRouter.POST("", ctr.Photo.Create) // 上传投稿（需登录）

			// 答题与评论
			photoRouter.POST("/:id/attempts", ctr.Attempt.Submit) // 提交答案（需登录）
			photoRouter.POST("/:id/comments", ctr.Comment.Create) // 发表评论（需登录）

		}

		// --- 我的奖品 ---
		apiRouter.GET("/users/me/prizes", ctr.Prize.MyPrizes)

		// --- 个人主页 ---
		apiRouter.GET("/users/:id", ctr.Auth.UserProfile)              // 访问他人首页
		apiRouter.GET("/users/:id/photos", ctr.Photo.UserPhotos)       // 个人主页-图片
		apiRouter.GET("/users/:id/attempts", ctr.Attempt.UserAttempt)  // 个人主页-答题
		apiRouter.GET("/users/:id/comments", ctr.Comment.UserComments) // 个人主页-评论

		// --- 管理员接口 ---
		adminRouter := apiRouter.Group("/admin")
		adminRouter.Use(middleware.CheckRole(1))
		{
			adminRouter.GET("/photos/pending", ctr.Admin.PendingPhotos)
			adminRouter.PUT("/photos/:id/review", ctr.Admin.ReviewPhoto)
			adminRouter.GET("/attempts/pending", ctr.Admin.PendingAttempts)
			adminRouter.PUT("/attempts/:id/review", ctr.Admin.ReviewAttempt)
			adminRouter.PUT("/prizes/:id/claim", ctr.Admin.ClaimPrize)
			adminRouter.GET("/comments/pending", ctr.Admin.PendingComments)
			adminRouter.PUT("/comments/:id/review", ctr.Admin.ReviewComment)
		}

		// --- 消息通知 ---
		messageRouter := apiRouter.Group("/messages")
		messageRouter.Use(middleware.CheckRole(0))
		{
			messageRouter.GET("", ctr.Message.ListMyMessages)
			messageRouter.GET("/unread-count", ctr.Message.GetUnreadCount)
			messageRouter.PUT("/:id/read", ctr.Message.MarkAsRead)
		}
	}
}
