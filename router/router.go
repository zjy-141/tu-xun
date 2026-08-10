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

		// --- 测试登录 ---
		apiRouter.GET("/test/login", ctr.Test.Login)

		// 用户接口
		userRouter := apiRouter.Group("/user")
		{
			// userRouter.GET("/login", ctr.User.UserLogin) // 接口修改6：登录重定向由前端发起，后端不再提供此接口
			userRouter.GET("/logincallback", ctr.User.LoginCallback)
			userRouter.DELETE("/logout", ctr.User.UserLogout)

			userRouter.Use(middleware.CheckRole(1))
			userRouter.GET("/info", ctr.User.UserInfo)
			userRouter.PUT("/nickname", ctr.User.UpdateNickname)
			userRouter.PUT("/avatar", ctr.User.UploadAvatar)
			userRouter.POST("/avatar", ctr.User.UploadAvatar)
		}

		// --- 活动（公开） ---
		apiRouter.GET("/activity", ctr.Activity.List)

		// --- 图寻题目 ---
		photoRouter := apiRouter.Group("/photos")
		{
			// 公开接口
			photoRouter.GET("", ctr.Photo.List)
			photoRouter.GET("/:id", ctr.Photo.Detail)
			photoRouter.GET("/:id/comments", ctr.Photo.PhotoComments)
			photoRouter.GET("/:id/solves", ctr.Photo.PhotoSolves)

			// 需登录
			photoRouter.Use(middleware.CheckRole(1))
			photoRouter.POST("", ctr.Photo.Create)
			photoRouter.POST("/:id/attempts", ctr.Attempt.Submit)
			photoRouter.POST("/:id/comments", ctr.Comment.Create)
			photoRouter.PUT("/:id/like", ctr.Like.SetPhotoLike)
			photoRouter.GET("/:id/attempts/user", ctr.Photo.PhotoAttemptsUser)
			photoRouter.GET("/user", ctr.Photo.ListUser)
			photoRouter.GET("/user/:id", ctr.Photo.DetailUser)
		}

		// --- 作答 ---
		attemptRouter := apiRouter.Group("/attempts")
		attemptRouter.Use(middleware.CheckRole(1))
		{
			attemptRouter.GET("/user", ctr.Attempt.ListUser)
		}

		// --- 破解记录点赞 ---
		solvesRouter := apiRouter.Group("/solves")
		solvesRouter.Use(middleware.CheckRole(1))
		{
			solvesRouter.PUT("/:id/like", ctr.Like.SetSolveLike)
		}

		// --- 评论 ---
		commentRouter := apiRouter.Group("/comments")
		commentRouter.Use(middleware.CheckRole(1))
		{
			commentRouter.DELETE("/:id", ctr.Comment.Delete)
			commentRouter.PUT("/:id/like", ctr.Like.SetCommentLike)
		}

		// --- 积分 ---
		scoreRouter := apiRouter.Group("/score")
		scoreRouter.Use(middleware.CheckRole(1))
		{
			scoreRouter.GET("/logs", ctr.Score.MyScoreLog)
		}

		// --- 奖品 ---
		goodRouter := apiRouter.Group("/goods")
		goodRouter.Use(middleware.CheckRole(1))
		{
			goodRouter.GET("", ctr.Good.List)
		}

		// --- 兑换 ---
		exchangeRouter := apiRouter.Group("/exchange")
		exchangeRouter.Use(middleware.CheckRole(1))
		{
			exchangeRouter.POST("", ctr.Exchange.Claim)
			exchangeRouter.GET("", ctr.Exchange.List)
		}

		// --- 通知（公告） ---
		announcementRouter := apiRouter.Group("/announcements")
		announcementRouter.Use(middleware.CheckRole(1))
		{
			announcementRouter.GET("", ctr.Announcement.List)
			announcementRouter.GET("/:id", ctr.Announcement.Detail)
		}

		// --- 互动消息 ---
		notificationRouter := apiRouter.Group("/notifications")
		notificationRouter.Use(middleware.CheckRole(1))
		{
			notificationRouter.GET("", ctr.Message.ListInteractionMessages)
			notificationRouter.PUT("/:id/read", ctr.Message.MarkInteractionRead)
			notificationRouter.PUT("/read-all", ctr.Message.MarkAllInteractionRead)
		}

		// --- 反馈 ---
		feedbackRouter := apiRouter.Group("/feedback")
		feedbackRouter.Use(middleware.CheckRole(1))
		{
			feedbackRouter.POST("", ctr.Feedback.Create)
		}

		// --- 内容位（公开） ---
		apiRouter.GET("/contents/:key", ctr.ContentBlock.Get)

		// --- 管理员接口 ---
		adminRouter := apiRouter.Group("/admin")
		adminRouter.Use(middleware.CheckRole(2))
		{
			// 题目管理与审核
			adminRouter.GET("/photos", ctr.Admin.ListPhotos)
			adminRouter.POST("/photos", ctr.Admin.CreatePhoto)
			adminRouter.PUT("/photos/:id", ctr.Admin.UpdatePhoto)
			adminRouter.PUT("/photos/:id/review", ctr.Admin.ReviewPhoto)

			// 作答审核
			adminRouter.GET("/attempts", ctr.Admin.ListAttempts)
			adminRouter.PUT("/attempts/:id/review", ctr.Admin.ReviewAttempt)

			// 评论审核
			adminRouter.GET("/comments", ctr.Admin.ListComments)
			adminRouter.PUT("/comments/:id/review", ctr.Admin.ReviewComment)

			// 活动管理
			adminRouter.GET("/activity", ctr.AdminActivity.List)
			adminRouter.POST("/activity", ctr.AdminActivity.Create)
			adminRouter.PUT("/activity/:id", ctr.AdminActivity.Update)

			// 奖品管理
			adminRouter.GET("/goods", ctr.AdminGood.List)
			adminRouter.POST("/goods", ctr.AdminGood.Create)
			adminRouter.PUT("/goods/:id", ctr.AdminGood.Update)
			adminRouter.DELETE("/goods/:id", ctr.AdminGood.Delete)

			// 兑换管理
			adminRouter.GET("/exchange", ctr.AdminExchange.List)
			adminRouter.PUT("/exchange/:id/verify", ctr.AdminExchange.Verify)

			// 反馈管理
			adminRouter.GET("/feedback", ctr.Feedback.List)
			adminRouter.GET("/feedback/:id", ctr.Feedback.Detail)
			adminRouter.PUT("/feedback/:id", ctr.Feedback.Review)

			// 通知管理
			adminRouter.GET("/announcements", ctr.Announcement.AdminList)
			adminRouter.GET("/announcements/:id", ctr.Announcement.AdminDetail)
			adminRouter.POST("/announcements", ctr.Announcement.AdminCreate)
			adminRouter.PUT("/announcements/:id", ctr.Announcement.AdminUpdate)
			adminRouter.DELETE("/announcements/:id", ctr.Announcement.AdminDelete)

			// 内容位管理
			adminRouter.PUT("/contents/:key", ctr.ContentBlock.AdminUpdate)

			// 工作台统计
			adminRouter.GET("/stats", ctr.Admin.GetStats)

			// 超级管理员专属 (Level >= 3)
			superAdminRouter := adminRouter.Group("")
			superAdminRouter.Use(middleware.CheckRole(3))
			{
				superAdminRouter.GET("/users", ctr.Admin.ListUsers)
				superAdminRouter.PUT("/users/:id/status", ctr.Admin.SetUserStatus)
				superAdminRouter.PUT("/users/:id/level", ctr.Admin.UpdateAdminLevel)
			}
		}
	}
}
