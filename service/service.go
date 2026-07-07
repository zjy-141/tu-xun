package service

var OSSClient *OSS

type Service struct {
	UserSvc
	ActivitySvc
	PhotoSvc
	AttemptSvc
	AdminSvc
	ScoreSvc
	GoodSvc
	AdminGoodSvc
	AdminExchangeSvc
	ExchangeSvc
	CommentSvc
	MessageSvc
	LikeSvc
	OSS *OSS
}

// New 创建并返回聚合所有子服务的 Service 实例，同时初始化 OSS 客户端
func New() *Service {
	service := &Service{}
	service.OSS = NewOSS()
	OSSClient = service.OSS
	return service
}
