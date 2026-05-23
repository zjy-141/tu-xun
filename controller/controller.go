package controller

type Controller struct {
	Auth
	Photo
	Attempt
	Admin
	Prize
	Comment
}

func New() *Controller {
	Controller := &Controller{}
	return Controller
}
