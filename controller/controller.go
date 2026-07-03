package controller

type Controller struct {
	User
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
