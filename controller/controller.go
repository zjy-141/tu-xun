package controller

type Controller struct {
	Auth
	Photo
	Attempt
	Admin
	Prize
	Comment
	Message
}

func New() *Controller {
	Controller := &Controller{}
	return Controller
}
