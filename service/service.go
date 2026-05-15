package service

type Service struct {
	Auth
	Photo
	Attempt
	Admin
	Prize
	Story
}

func New() *Service {
	service := &Service{}
	return service
}
