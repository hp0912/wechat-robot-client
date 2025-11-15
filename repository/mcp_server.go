package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"wechat-robot-client/model"
)

type MCPServer struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewMCPServerRepo(ctx context.Context, db *gorm.DB) *MCPServer {
	return &MCPServer{Ctx: ctx, DB: db}
}

func (respo *MCPServer) Create(server *model.MCPServer) error {
	return respo.DB.WithContext(respo.Ctx).Create(server).Error
}

func (respo *MCPServer) Update(server *model.MCPServer) error {
	return respo.DB.WithContext(respo.Ctx).Updates(server).Error
}

func (respo *MCPServer) Delete(id uint64) error {
	return respo.DB.WithContext(respo.Ctx).Unscoped().Delete(&model.MCPServer{}, id).Error
}

func (respo *MCPServer) FindByID(id uint64) (*model.MCPServer, error) {
	var server model.MCPServer
	err := respo.DB.WithContext(respo.Ctx).First(&server, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (respo *MCPServer) FindByName(name string) (*model.MCPServer, error) {
	var server model.MCPServer
	err := respo.DB.WithContext(respo.Ctx).Where("name = ?", name).First(&server).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (respo *MCPServer) FindAll() ([]*model.MCPServer, error) {
	var servers []*model.MCPServer
	err := respo.DB.WithContext(respo.Ctx).Order("priority DESC, id ASC").Find(&servers).Error
	return servers, err
}

func (respo *MCPServer) FindEnabled() ([]*model.MCPServer, error) {
	var servers []*model.MCPServer
	enabled := true
	err := respo.DB.WithContext(respo.Ctx).Where("enabled = ?", enabled).
		Order("priority DESC, id ASC").
		Find(&servers).Error
	return servers, err
}

func (respo *MCPServer) FindByTransport(transport model.MCPTransportType) ([]*model.MCPServer, error) {
	var servers []*model.MCPServer
	err := respo.DB.WithContext(respo.Ctx).Where("transport = ?", transport).
		Order("priority DESC, id ASC").
		Find(&servers).Error
	return servers, err
}

func (respo *MCPServer) FindByTag(tag string) ([]*model.MCPServer, error) {
	var servers []*model.MCPServer
	// MySQL的JSON_CONTAINS函数
	err := respo.DB.WithContext(respo.Ctx).Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, tag)).
		Order("priority DESC, id ASC").
		Find(&servers).Error
	return servers, err
}

// FindEnabledByTransport 根据传输类型查询已启用的MCP服务器配置
func (respo *MCPServer) FindEnabledByTransport(transport model.MCPTransportType) ([]*model.MCPServer, error) {
	var servers []*model.MCPServer
	enabled := true
	err := respo.DB.WithContext(respo.Ctx).Where("enabled = ? AND transport = ?", enabled, transport).
		Order("priority DESC, id ASC").
		Find(&servers).Error
	return servers, err
}

// UpdateEnabled 更新启用状态
func (respo *MCPServer) UpdateEnabled(id uint64, enabled bool) error {
	return respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Where("id = ?", id).
		Update("enabled", enabled).Error
}

// UpdatePriority 更新优先级
func (respo *MCPServer) UpdatePriority(id uint64, priority int) error {
	return respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Where("id = ?", id).
		Update("priority", priority).Error
}

// IncrementConnectionCount 增加连接计数
func (respo *MCPServer) IncrementConnectionCount(id uint64) error {
	return respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Where("id = ?", id).
		UpdateColumn("connection_count", gorm.Expr("connection_count + 1")).Error
}

// IncrementErrorCount 增加错误计数
func (respo *MCPServer) IncrementErrorCount(id uint64) error {
	return respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Where("id = ?", id).
		UpdateColumn("error_count", gorm.Expr("error_count + 1")).Error
}

// UpdateConnectionSuccess 更新连接成功状态
func (respo *MCPServer) UpdateConnectionSuccess(id uint64) error {
	now := time.Now()
	return respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_connected_at": &now,
			"last_error":        "",
			"connection_count":  gorm.Expr("connection_count + 1"),
		}).Error
}

// UpdateConnectionError 更新连接错误状态
func (respo *MCPServer) UpdateConnectionError(id uint64, errMsg string) error {
	return respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_error":       errMsg,
			"connection_count": gorm.Expr("connection_count + 1"),
			"error_count":      gorm.Expr("error_count + 1"),
		}).Error
}

// ResetErrorCount 重置错误计数
func (respo *MCPServer) ResetErrorCount(id uint64) error {
	return respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"error_count": 0,
			"last_error":  "",
		}).Error
}

// CountByTransport 统计各传输类型的服务器数量
func (respo *MCPServer) CountByTransport() (map[model.MCPTransportType]int64, error) {
	type Result struct {
		Transport model.MCPTransportType
		Count     int64
	}

	var results []Result
	err := respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Select("transport, COUNT(*) as count").
		Group("transport").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	countMap := make(map[model.MCPTransportType]int64)
	for _, result := range results {
		countMap[result.Transport] = result.Count
	}

	return countMap, nil
}

// FindWithHighErrorRate 查询高错误率的服务器（错误率 > 阈值）
func (respo *MCPServer) FindWithHighErrorRate(errorRateThreshold float64) ([]*model.MCPServer, error) {
	var servers []*model.MCPServer

	// 计算错误率：error_count / (connection_count + error_count)
	// 只查询至少有10次连接的服务器
	err := respo.DB.WithContext(respo.Ctx).Where("connection_count + error_count >= 10").
		Where("error_count::float / (connection_count + error_count) > ?", errorRateThreshold).
		Order("error_count DESC").
		Find(&servers).Error

	return servers, err
}

// FindNotConnectedRecently 查询最近未连接的服务器
func (respo *MCPServer) FindNotConnectedRecently(duration time.Duration) ([]*model.MCPServer, error) {
	var servers []*model.MCPServer
	enabled := true
	cutoffTime := time.Now().Add(-duration)

	err := respo.DB.WithContext(respo.Ctx).Where("enabled = ?", enabled).
		Where("last_connected_at IS NULL OR last_connected_at < ?", cutoffTime).
		Order("last_connected_at ASC").
		Find(&servers).Error

	return servers, err
}

// BatchUpdateEnabled 批量更新启用状态
func (respo *MCPServer) BatchUpdateEnabled(ids []uint64, enabled bool) error {
	return respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).
		Where("id IN ?", ids).
		Update("enabled", enabled).Error
}

// Exists 检查配置是否存在
func (respo *MCPServer) Exists(id uint64) (bool, error) {
	var count int64
	err := respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByName 检查名称是否已存在
func (respo *MCPServer) ExistsByName(name string) (bool, error) {
	var count int64
	err := respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// GetTotalCount 获取总数
func (respo *MCPServer) GetTotalCount() (int64, error) {
	var count int64
	err := respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).Count(&count).Error
	return count, err
}

// GetEnabledCount 获取已启用的数量
func (respo *MCPServer) GetEnabledCount() (int64, error) {
	var count int64
	enabled := true
	err := respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).Where("enabled = ?", enabled).Count(&count).Error
	return count, err
}

// FindWithPagination 分页查询
func (respo *MCPServer) FindWithPagination(page, pageSize int) ([]*model.MCPServer, int64, error) {
	var servers []*model.MCPServer
	var total int64

	// 查询总数
	if err := respo.DB.WithContext(respo.Ctx).Model(&model.MCPServer{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := respo.DB.WithContext(respo.Ctx).Order("priority DESC, id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&servers).Error

	return servers, total, err
}

// Transaction 执行事务
func (respo *MCPServer) Transaction(fn func(*gorm.DB) error) error {
	return respo.DB.WithContext(respo.Ctx).Transaction(fn)
}
