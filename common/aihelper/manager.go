package aihelper

import (
	"context"
	"sync"
)

//采用单例模式管理用户-会话-AIHelper的映射关系，实现实例缓存和生命周期控制。

var ctx = context.Background()

type AIHelperManager struct {
	helpers map[string]map[string]*AIHelper
	mu      sync.RWMutex
}

func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]map[string]*AIHelper),
	}
}

func (m *AIHelperManager) GetOrCreateAIHelper(userName string, sessionID string, modelType string, config map[string]interface{}) (*AIHelper, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	//检验用户存不存在，获取用户的会话
	userHelpers, exists := m.helpers[userName]

	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[userName] = userHelpers
	}

	//检验会话存不存在，存在就成功GET到aihelper
	helper, exists := userHelpers[sessionID]
	if exists {
		return helper, nil
	}

	//会话不存在，创建
	factory := GetGlobalFactory()
	helper, err := factory.CreateAIHelper(ctx, modelType, sessionID, config)

	if err != nil {
		return nil, err
	}

	userHelpers[sessionID] = helper
	return helper, nil
}

// 获取指定用户指定对话的aihelper
func (m *AIHelperManager) GetAIHelper(userName, sessionID string) (*AIHelper, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	userHelpers, exists := m.helpers[userName]
	if !exists {
		return nil, false
	}

	helper, exists := userHelpers[sessionID]
	return helper, exists
}

// 删除该会话的aihelper
func (m *AIHelperManager) RemoveAIHelper(userName, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	userHelpers, exists := m.helpers[userName]
	if !exists {
		return
	}

	delete(userHelpers, sessionID)
	if len(userHelpers) == 0 {
		delete(m.helpers, userName)
	}
}

// 获取指定用户所有会话id
func (m *AIHelperManager) GetUserSessions(userName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return []string{}
	}

	sessionIDs := make([]string, 0, len(userHelpers))

	for sessionID := range userHelpers {
		sessionIDs = append(sessionIDs, sessionID)
	}
	return sessionIDs
}

// 全局管理器
var globalManager *AIHelperManager
var once sync.Once //只实例化一次

func GetGlobalManager() *AIHelperManager {
	once.Do(func() {
		globalManager = NewAIHelperManager()
	})
	return globalManager
}
