package session

import (
	"GopherAI/model"
	"context"
)

var ctx  = context.Background()

func GetUserSessionsByUserName(userName string)([]model.SessionInfo,error)