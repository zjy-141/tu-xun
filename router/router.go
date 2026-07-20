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
		testRouter := apiRouter.Group("/tenzor/tiaozhan/test")
		{
			testRouter.GET("/login", ctr.Test.Login)
		}

		// 用户接口
		userRouter := apiRouter.Group("/user")
		{
			userRouter.GET("/login", ctr.User.UserLogin)
			userRouter.GET("/logincallback", ctr.User.LoginCallback)
			userRouter.DELETE("/logout", ctr.User.UserLogout)

			userRouter.Use(middleware.CheckRole(1))
			userRouter.GET("/info", ctr.User.UserInfo)
			userRouter.PUT("/info", ctr.User.UpdateUserInfo)
			userRouter.PUT("/avatar", ctr.User.UploadAvatar)
		}
		// --- 活动 ---
		activityRouter := apiRouter.Group("/activity")
		{
			activityRouter.GET("/current", ctr.Activity.CurrentActivity)
			activityRouter.GET("/history", ctr.Activity.HistoryActivity)
		}
		// --- 图片（图寻题目） ---
		photoRouter := apiRouter.Group("/photos")
		{
			photoRouter.GET("/list", ctr.Photo.List)
			photoRouter.GET("/:id", ctr.Photo.Detail)
			photoRouter.GET("/:id/image", ctr.Photo.GetImageStream)
			photoRouter.GET("/:id/download", ctr.Photo.Download)
			photoRouter.GET("/:id/comments", ctr.Photo.PhotoComments)
			photoRouter.GET("/:id/attempts", ctr.Photo.PhotoAttempts)

			photoRouter.Use(middleware.CheckRole(1))
			photoRouter.GET("/:id/attempts/user", ctr.Photo.PhotoAttemptsUser)

			photoRouter.POST("/:id/attempts", ctr.Attempt.Submit)        // 提交答题
			photoRouter.POST("/:id/comments", ctr.Comment.Create)        // 发表评论
			photoRouter.POST("", ctr.Photo.Create)                       // 上传投稿
			photoRouter.GET("/user", ctr.Photo.ListUser)                 // 获取该用户投稿的图片列表
			photoRouter.GET("/review/:id", ctr.Photo.DetailUser)         // 获取该用户投稿的图片详情

			// 点赞
			photoRouter.POST("/:id/like", ctr.Like.TogglePhotoLike)
			photoRouter.GET("/:id/like", ctr.Like.GetPhotoLikeStatus)
		}

		// --- 答题记录 ---
		attemptLikeRouter := apiRouter.Group("/attempts")
		attemptLikeRouter.Use(middleware.CheckRole(1))
		{
			attemptLikeRouter.POST("/:id/like", ctr.Like.ToggleAttemptLike)
			attemptLikeRouter.GET("/:id/like", ctr.Like.GetAttemptLikeStatus)

			attemptLikeRouter.GET("/user", ctr.Attempt.ListUser)
		}

		// --- 评论 ---
		commentRouter := apiRouter.Group("/comments")
		commentRouter.Use(middleware.CheckRole(1))
		{
			commentRouter.DELETE("/:id", ctr.Comment.Delete)
			commentRouter.POST("/:id/like", ctr.Like.ToggleCommentLike)
			commentRouter.GET("/:id/like", ctr.Like.GetCommentLikeStatus)
		}

		// --- 用户主页/积分 ---
		scoreRouter := apiRouter.Group("/score")
		scoreRouter.Use(middleware.CheckRole(1))
		{
			scoreRouter.GET("", ctr.Score.MyScore)
			scoreRouter.GET("/logs", ctr.Score.MyScoreLog)
		}
		// --- 奖品 ---
		goodRouter := apiRouter.Group("/goods")
		goodRouter.Use(middleware.CheckRole(1))
		{
			goodRouter.GET("/list", ctr.Good.List)
			goodRouter.GET("/:id", ctr.Good.Detail)
		}
		// --- 兑换奖品 ---
		exchangeRouter := apiRouter.Group("/exchange")
		exchangeRouter.Use(middleware.CheckRole(1))
		{
			exchangeRouter.POST("/claim", ctr.Exchange.Claim)
			exchangeRouter.GET("/list", ctr.Exchange.List)
		}
		// --- 消息通知 ---
		messageRouter := apiRouter.Group("/messages")
		messageRouter.Use(middleware.CheckRole(1))
		{
			messageRouter.GET("/list", ctr.Message.List)
			messageRouter.GET("/:id", ctr.Message.Detail)
			messageRouter.GET("/unread-count", ctr.Message.GetUnreadCount)
			messageRouter.PUT("/:id/read", ctr.Message.MarkAsRead)
			messageRouter.GET("/notice", ctr.Message.Notice)
		}
		// --- 反馈（用户提交） ---
		feedbackRouter := apiRouter.Group("/feedback")
		feedbackRouter.Use(middleware.CheckRole(1))
		{
			feedbackRouter.POST("", ctr.Feedback.Create)
		}
		// --- 管理员接口 ---
		adminRouter := apiRouter.Group("/admin")
		adminRouter.Use(middleware.CheckRole(2))
		{
			adminRouter.GET("/photos/pending", ctr.Admin.PendingPhotos)
			adminRouter.PUT("/photos/:id/review", ctr.Admin.ReviewPhoto)
			adminRouter.GET("/attempts/pending", ctr.Admin.PendingAttempts)
			adminRouter.PUT("/attempts/:id/review", ctr.Admin.ReviewAttempt)
			adminRouter.GET("/comments/pending", ctr.Admin.PendingComments)
			adminRouter.PUT("/comments/:id/review", ctr.Admin.ReviewComment)

			// 管理员活动管理
			activityRouter := adminRouter.Group("/activity")
			{
				activityRouter.GET("/list", ctr.AdminActivity.List)
				activityRouter.GET("/:id", ctr.AdminActivity.Detail)
				activityRouter.POST("/create", ctr.AdminActivity.Create)
				activityRouter.POST("/update", ctr.AdminActivity.Update)
				activityRouter.POST("/notice", ctr.AdminActivity.Notice)
			}

			// 管理员商品管理
			admingoodRouter := adminRouter.Group("/goods")
			{
				admingoodRouter.GET("/list", ctr.AdminGood.List)
				admingoodRouter.GET("/:id", ctr.AdminGood.Detail)
				admingoodRouter.POST("/new", ctr.AdminGood.Create)
				admingoodRouter.PUT("/:id", ctr.AdminGood.Update)
				admingoodRouter.DELETE("/:id", ctr.AdminGood.Delete)
				admingoodRouter.PUT("/:id/status", ctr.AdminGood.Status)
				admingoodRouter.PUT("/:id/stock", ctr.AdminGood.Stock)
			}
			// 管理员兑换管理
			adminexchangeRouter := adminRouter.Group("/exchange")
			{
				adminexchangeRouter.GET("/list", ctr.AdminExchange.List)
				adminexchangeRouter.POST("/verify", ctr.AdminExchange.Verify)
			}
			// 管理员反馈管理
			adminRouter.GET("/feedback/list", ctr.Feedback.List)
			adminRouter.GET("/feedback/:id", ctr.Feedback.Detail)
			adminRouter.PUT("/feedback/:id", ctr.Feedback.Review)

			// 管理员用户搜索
			adminRouter.GET("/users", ctr.Admin.SearchUsers)

			// 高级管理员专属 (Level >= 3)
			superAdminRouter := adminRouter.Group("")
			superAdminRouter.Use(middleware.CheckRole(3))
			{
				superAdminRouter.GET("/user", ctr.Admin.UserList)
				superAdminRouter.PUT("/level", ctr.Admin.UpdateAdminLevel)
			}
		}
	}
}
