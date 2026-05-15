package controller

type Controller struct {
	Auth
	Photo
	Attempt
	Admin
	Prize
	Story
}

func New() *Controller {
	Controller := &Controller{}
	return Controller
}
