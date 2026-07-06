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
	Good
	Exchange
}

func New() *Controller {
	Controller := &Controller{}
	return Controller
}
