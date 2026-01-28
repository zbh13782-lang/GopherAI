package aihelper

import (
	"context"
	"fmt"
	"sync"
)

type ModelCreator func(ctx context.Context, config map[string]interface{}) (AIModel, error)

type AIModelFactory struct {
	creators map[string]ModelCreator //1：openai，2：ollama
}

var (
	globalFactory *AIModelFactory
	factoryonce   sync.Once //确保只初始化globalfactory一次
)

func GetGlobalFactory() *AIModelFactory {
	factoryonce.Do(func() {
		globalFactory = &AIModelFactory{
			creators: make(map[string]ModelCreator),
		}

		globalFactory.registerCreators()
	})
	return globalFactory
}

// 注册模型
func (f *AIModelFactory) registerCreators() {
	f.creators["1"] = func(ctx context.Context, config map[string]interface{}) (AIModel, error) {
		return NewOpenAIModel(ctx)
	}

	f.creators["2"] = func(ctx context.Context, config map[string]interface{}) (AIModel, error) {
		baseURL, _ := config["baseURL"].(string)
		modelName, ok := config["modelName"].(string)

		if !ok {
			return nil, fmt.Errorf("Ollama model requires modelname")
		}
		return NewOllamaModel(ctx, baseURL, modelName)
	}
}

func (f *AIModelFactory) CreateAIModel(ctx context.Context, modelType string, config map[string]interface{}) (AIModel, error) {
	creator, ok := f.creators[modelType]

	if !ok {
		return nil, fmt.Errorf("unsupported model type %s", modelType)
	}

	return creator(ctx, config)
}

func (f *AIModelFactory) CreateAIHelper(ctx context.Context, modelType string, SessionID string, config map[string]interface{}) (*AIHelper, error) {
	model, err := f.CreateAIModel(ctx, modelType, config)
	if err != nil {
		return nil, err
	}

	return NewAIHelper(model, SessionID), nil
}
// 可拓展注册
func (f *AIModelFactory) RegisterModel(modelType string, creator ModelCreator) {
	f.creators[modelType] = creator
}
