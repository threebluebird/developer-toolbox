package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"developer-toolbox/backend/config"
	appLogger "developer-toolbox/backend/logger"
	"developer-toolbox/backend/models"
	"developer-toolbox/backend/registry"
	"developer-toolbox/backend/services"
	"developer-toolbox/backend/storage"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the root application container for the toolbox.
type App struct {
	ctx             context.Context
	name            string
	toolService     *services.ToolService
	historyService  *services.HistoryService
	favoriteService *services.FavoriteService
	configService   *services.ConfigService
	mu              sync.RWMutex
	logger          appLogger.Logger
	fileLogger      *appLogger.FileLogger
}

// NewApp 负责组装应用依赖，是后端唯一的“组合根”。
// 设置文件始终从默认 data 目录读取；若用户配置了自定义目录，
// 收藏和历史记录会切换到该目录，保证下次启动仍能找到路径配置。
func NewApp() *App {
	logPath := filepath.Join(config.StorageDir, "developer-toolbox.log")
	fileLogger, logErr := appLogger.NewFileLogger(logPath, appLogger.InfoLevel)
	var applicationLogger appLogger.Logger
	if logErr == nil {
		applicationLogger = fileLogger
	} else {
		applicationLogger = appLogger.New(os.Stderr, appLogger.InfoLevel)
	}
	toolRegistry := registry.CreateDefaultRegistry()
	toolService := services.NewToolService(toolRegistry)
	fileStorage := storage.NewFileStorage(config.StorageDir)
	configService := services.NewConfigService(storage.NewConfigRepository(fileStorage))
	settings, err := configService.Load()
	if err != nil {
		settings = config.DefaultSettings()
	}
	toolService.SetContextFactory(newToolContextFactory(settings))
	if err == nil && strings.TrimSpace(settings.StoragePath) != "" {
		fileStorage = storage.NewFileStorage(settings.StoragePath)
	}
	historyService := services.NewHistoryService(storage.NewHistoryRepository(fileStorage))
	favoriteService := services.NewFavoriteService(storage.NewFavoriteRepository(fileStorage))

	return &App{
		name:            config.AppName,
		toolService:     toolService,
		historyService:  historyService,
		favoriteService: favoriteService,
		configService:   configService,
		logger:          applicationLogger,
		fileLogger:      fileLogger,
	}
}

func newToolContextFactory(settings models.Settings) func(context.Context) models.ToolContext {
	workspace, err := os.Getwd()
	if err != nil {
		workspace = ""
	}
	return func(ctx context.Context) models.ToolContext {
		toolCtx := models.NewToolContext(ctx)
		toolCtx.Workspace = workspace
		toolCtx.Config["jsonIndent"] = settings.JSONIndent
		toolCtx.UserData["theme"] = settings.Theme
		toolCtx.UserData["language"] = settings.Language
		return toolCtx
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("application started version=%s", config.AppVersion)
}

func (a *App) shutdown(context.Context) {
	a.logger.Info("application stopped")
	if a.fileLogger != nil {
		_ = a.fileLogger.Close()
	}
}

// Greet returns a greeting for the given name.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, welcome to %s!", name, a.name)
}

// Health returns a simple health payload so the UI can validate the app is alive.
func (a *App) Health() map[string]any {
	return map[string]any{
		"status":  "ok",
		"name":    a.name,
		"version": config.AppVersion,
	}
}

// ListTools returns all registered tools.
func (a *App) ListTools() []models.Tool {
	return a.toolService.ListTools()
}

func (a *App) SearchTools(query string) []models.Tool {
	return services.NewSearchService(a.toolService).Search(query)
}

func (a *App) SelectFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a file to hash",
	})
}

// ExecuteTool runs a tool by its ID using the provided input payload.
func (a *App) ExecuteTool(input models.ToolInput) models.ToolOutput {
	// 执行期间持有读锁，避免 SaveSettings 切换 Repository 时产生数据竞争。
	a.mu.RLock()
	defer a.mu.RUnlock()
	output := a.toolService.ExecuteToolContext(a.ctx, input)
	if output.Success {
		a.logger.Debug("tool completed id=%s", input.ToolID)
	} else if output.Error != nil {
		a.logger.Warn("tool failed id=%s code=%s", input.ToolID, output.Error.Code)
	}
	if output.Success {
		// 历史落盘失败不覆盖工具本身的成功结果，避免辅助功能影响主要操作。
		_ = a.historyService.Add(models.History{
			ID:     uuid.NewString(),
			ToolID: input.ToolID,
			UsedAt: time.Now().UTC().Format(time.RFC3339),
			Input:  safeSnapshot(input.Payload),
			Output: safeOutputSnapshot(output.Data),
		})
	}
	return output
}

const workflowSnapshotLimit = 64 * 1024

var sensitiveWorkflowKeys = map[string]bool{"password": true, "token": true, "authorization": true, "cookie": true, "apikey": true, "api-key": true, "secret": true}

func safeSnapshot(payload map[string]any) map[string]any {
	result := make(map[string]any, len(payload))
	for key, value := range payload {
		if sensitiveWorkflowKeys[strings.ToLower(strings.ReplaceAll(key, "_", ""))] {
			result[key] = "••••••••"
		} else {
			result[key] = redactValue(value)
		}
	}
	data, _ := json.Marshal(result)
	if len(data) > workflowSnapshotLimit {
		return map[string]any{"preview": string(data[:workflowSnapshotLimit]), "truncated": true}
	}
	return result
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return safeSnapshot(typed)
	case map[string]string:
		converted := make(map[string]any, len(typed))
		for key, value := range typed {
			converted[key] = value
		}
		return safeSnapshot(converted)
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = redactValue(item)
		}
		return result
	default:
		return value
	}
}

func safeOutputSnapshot(value any) any {
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	if len(data) > workflowSnapshotLimit {
		return map[string]any{"preview": string(data[:workflowSnapshotLimit]), "truncated": true}
	}
	return value
}

func (a *App) GetHistory() ([]models.History, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.historyService.List()
}

func (a *App) ClearHistory() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.historyService.Clear()
}

func (a *App) AddFavorite(toolID string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.favoriteService.Add(toolID)
}

func (a *App) RemoveFavorite(toolID string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.favoriteService.Remove(toolID)
}

func (a *App) AddFavoriteItem(item models.Favorite) (models.Favorite, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	item.ID = uuid.NewString()
	if item.Kind == "" {
		item.Kind = "input"
	}
	item.AddedAt = time.Now().UTC().Format(time.RFC3339)
	item.Payload = safeSnapshot(item.Payload)
	return item, a.favoriteService.AddItem(item)
}

func (a *App) RemoveFavoriteItem(id string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.favoriteService.RemoveItem(id)
}

func (a *App) ListFavorites() ([]models.Favorite, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.favoriteService.List()
}

func (a *App) LoadSettings() (models.Settings, error) {
	settings, err := a.configService.Load()
	if err != nil {
		return config.DefaultSettings(), err
	}
	return config.NormalizeSettings(settings), nil
}

func (a *App) SaveSettings(settings models.Settings) error {
	// 先归一化旧版或非法配置，保证落盘值始终处于前端支持的范围内。
	settings = config.NormalizeSettings(settings)
	settings.StoragePath = strings.TrimSpace(settings.StoragePath)
	if settings.StoragePath != "" {
		absolute, err := filepath.Abs(settings.StoragePath)
		if err != nil {
			return err
		}
		settings.StoragePath = absolute
	}
	if err := a.configService.Save(settings); err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	// 保存成功后立即重建依赖数据目录的 Repository，使新路径无需重启即可生效。
	dataPath := config.StorageDir
	if settings.StoragePath != "" {
		dataPath = settings.StoragePath
	}
	fileStorage := storage.NewFileStorage(dataPath)
	a.historyService = services.NewHistoryService(storage.NewHistoryRepository(fileStorage))
	a.favoriteService = services.NewFavoriteService(storage.NewFavoriteRepository(fileStorage))
	a.toolService.SetContextFactory(newToolContextFactory(settings))
	return nil
}
