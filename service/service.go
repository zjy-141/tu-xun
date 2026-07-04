package service

var OSSClient *OSS

type Service struct {
	UserSvc
	ActivitySvc
	PhotoSvc
	AttemptSvc
	AdminSvc
	PrizeSvc
	CommentSvc
	MessageSvc
	LikeSvc
	OSS *OSS
}

func New() *Service {
	service := &Service{}
	service.OSS = NewOSS()
	OSSClient = service.OSS
	return service
}
