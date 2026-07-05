package controller

type Controller struct {
	User
	Activity
	Photo
	Attempt
	Admin
	Comment
	Message
	Like
	Score
}

func New() *Controller {
	Controller := &Controller{}
	return Controller
}
