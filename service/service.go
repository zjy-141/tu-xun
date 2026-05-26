package service

var OSSClient *OSS

type Service struct {
	Auth
	Photo
	Attempt
	Admin
	Prize
	Comment
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
