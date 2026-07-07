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
		// 用户接口
		userRouter := apiRouter.Group("/user")
		{
			userRouter.GET("/login", ctr.User.UserLogin)
			userRouter.GET("/logincallback", ctr.User.LoginCallback)
			userRouter.DELETE("/logout", ctr.User.UserLogout)

			userRouter.Use(middleware.CheckRole(1))
			userRouter.GET("/userinfo", ctr.User.UserInfo)
			userRouter.PUT("/userinfo", ctr.User.UpdateUserInfo)
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
			photoRouter.GET("/list", ctr.Photo.List)                  // 公共浏览
			photoRouter.GET("/:id", ctr.Photo.Detail)                 // 图片详情
			photoRouter.GET("/:id/image", ctr.Photo.GetImageStream)   // 图片展示（流式）
			photoRouter.GET("/:id/download", ctr.Photo.Download)      // 图片下载
			photoRouter.GET("/:id/comments", ctr.Photo.PhotoComments) // 图片评论列表
			photoRouter.GET("/:id/attempts", ctr.Photo.PhotoAttempts) // 图片答题列表
			photoRouter.Use(middleware.CheckRole(1))
			photoRouter.POST("", ctr.Photo.Create) // 上传投稿（需登录）

			// 点赞
			photoRouter.POST("/:id/like", ctr.Like.TogglePhotoLike)   // 切换图片点赞
			photoRouter.GET("/:id/like", ctr.Like.GetPhotoLikeStatus) // 图片点赞状态
		}

		// --- 答题记录 ---
		attemptLikeRouter := apiRouter.Group("/attempts")
		attemptLikeRouter.Use(middleware.CheckRole(1))
		{
			attemptLikeRouter.POST("/:id/attempts", ctr.Attempt.Submit) // 提交答案（需登录）
			attemptLikeRouter.POST("/:id/like", ctr.Like.ToggleAttemptLike)
			attemptLikeRouter.GET("/:id/like", ctr.Like.GetAttemptLikeStatus)
		}

		// --- 评论 ---
		commentRouter := apiRouter.Group("/comments")
		commentRouter.Use(middleware.CheckRole(1))
		{
			commentRouter.POST("/:id/comments", ctr.Comment.Create) // 发表评论（需登录）
			commentRouter.DELETE("/:id", ctr.Comment.Delete)        // 删除评论（需登录）
			commentRouter.POST("/:id/like", ctr.Like.ToggleCommentLike)
			commentRouter.GET("/:id/like", ctr.Like.GetCommentLikeStatus)
		}

		// --- 用户主页 ---
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
			exchangeRouter.GET("/list", ctr.Good.List)
			exchangeRouter.GET("/:id", ctr.Good.Detail)
		}
		// --- 消息通知 ---
		messageRouter := apiRouter.Group("/messages")
		messageRouter.Use(middleware.CheckRole(1))
		{
			messageRouter.GET("/list", ctr.Message.List)
			messageRouter.GET("/:id", ctr.Message.Detail)
			messageRouter.GET("/unread-count", ctr.Message.GetUnreadCount)
			messageRouter.PUT("/:id/read", ctr.Message.MarkAsRead)
		}
		// --- 反馈 ---
		feedbackRouter := apiRouter.Group("/feedback")
		feedbackRouter.Use(middleware.CheckRole(1))
		{
			feedbackRouter.POST("", ctr.Message.FeedBack)
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

			adminRouter.POST("/notice", ctr.Admin.Announcement)

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

			adminexchangeRouter := adminRouter.Group("/exchange")
			{
				adminexchangeRouter.GET("/list", ctr.AdminExchange.List)
				adminexchangeRouter.POST("/verify", ctr.AdminExchange.Verify)
			}
			// 高级管理员专属 (Level >= 3)
			superAdminRouter := adminRouter.Group("")
			superAdminRouter.Use(middleware.CheckRole(3))
			{
				superAdminRouter.PUT("/admins/:id/level", ctr.Admin.UpdateAdminLevel)
			}
		}

		// // --- 会话（微信风格聊天） ---
		// conversationRouter := apiRouter.Group("/conversations")
		// conversationRouter.Use(middleware.CheckRole(0))
		// {
		// 	conversationRouter.GET("", ctr.Message.ListConversations)
		// 	conversationRouter.GET("/:id", ctr.Message.GetConversation)
		// 	conversationRouter.POST("/:id", ctr.Message.SendChatMessage)
		// }
	}
}
