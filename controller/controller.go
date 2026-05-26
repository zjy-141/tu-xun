package controller

type Controller struct {
	Auth
	Photo
	Attempt
	Admin
	Prize
	Comment
	Message
	Like
}

func New() *Controller {
	Controller := &Controller{}
	return Controller
}
