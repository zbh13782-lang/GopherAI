package aihelper

import (
	"GopherAI/common/rabbitmq"
	"GopherAI/model"
	"GopherAI/utils"
	"context"
	"sync"
)

type AIHelper struct {
	model     AIModel
	messages  []*model.Message //历史消息
	mu        sync.RWMutex
	SessionID string
	saveFunc  func(*model.Message) (*model.Message, error) // 异步发送到MQ
}

func NewAIHelper(model_ AIModel, SessionID string) *AIHelper {
	return &AIHelper{
		model:    model_,
		messages: make([]*model.Message, 0),

		saveFunc: func(msg *model.Message) (*model.Message, error) {
			if rabbitmq.RMQMessage == nil {
				// RabbitMQ未初始化，跳过发布
				return msg, nil
			}
			data := rabbitmq.GenerateMessageMQParam(msg.SessionID, msg.Content, msg.UserName, msg.IsUser)
			err := rabbitmq.RMQMessage.Publish(data)
			return msg, err
		},
		SessionID: SessionID,
	}
}

func (a *AIHelper) AddMessage(Content string, UserName string, IsUser bool, Save bool) {
	userMsg := model.Message{
		SessionID: a.SessionID,
		Content:   Content,
		UserName:  UserName,
		IsUser:    IsUser,
	}

	a.messages = append(a.messages, &userMsg)

	if Save {
		a.saveFunc(&userMsg)
	}
}

// 设置saveFunc
func (a *AIHelper) SetSaveFunc(savefunc func(*model.Message) (*model.Message, error)) {
	a.saveFunc = savefunc
}

// 获取历史消息
func (a *AIHelper) GetMessages() []*model.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]*model.Message, len(a.messages))
	copy(out, a.messages)
	return out
}

// 同步生成
func (a *AIHelper) GenerateResponse(userName string, ctx context.Context, userQuestion string) (*model.Message, error) {

	a.AddMessage(userQuestion, userName, true, true)
	a.mu.RLock()

	//model.msg->schema.msg,让ai看
	messages := utils.ConvertToSchemaMessages(a.messages)
	a.mu.RUnlock()

	//调用模型(同步生成)
	schemaMsg, err := a.model.GenerateResponse(ctx, messages)
	if err != nil {
		return nil, err
	}

	//把ai回复的schame.msg格式转换成model.msg
	modelMsg := utils.ConvertToModelMessage(a.SessionID, userName, schemaMsg)
	a.AddMessage(modelMsg.Content, userName, false, true)
	return modelMsg, nil
}

// 流式生成
func (a *AIHelper) StreamResponse(userName string, ctx context.Context, cb StreamCallback, userQuestion string) (*model.Message, error) {
	a.AddMessage(userQuestion, userName, true, true)
	a.mu.RLock()
	messages := utils.ConvertToSchemaMessages(a.messages)
	a.mu.RUnlock()
	content, err := a.model.StreamResponse(ctx, messages, cb)
	if err != nil {
		return nil, err
	}

	modelMsg := &model.Message{
		SessionID: a.SessionID,
		UserName:  userName,
		Content:   content,
		IsUser:    false,
	}
	a.AddMessage(modelMsg.Content, userName, false, true)
	return modelMsg, nil
}

// GetModelType 获取模型类型
func (a *AIHelper) GetModelType() string {
	return a.model.GetModelType()
}
