package api

import (
	"log"
	"net/http"

	"sop-chat/pkg/sopchat"

	"github.com/gin-gonic/gin"
)

// handleGetAccountId 获取当前阿里云账号ID
func (s *Server) handleGetAccountId(c *gin.Context) {
	accountId, err := sopchat.GetAccountId(s.config.AccessKeyId, s.config.AccessKeySecret)
	if err != nil {
		log.Printf("Failed to get account ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to get account ID",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accountId": accountId,
	})
}

// handleGetSystemConfig 获取系统配置（语言、时区等）
func (s *Server) handleGetSystemConfig(c *gin.Context) {
	language := "zh"
	timeZone := "Asia/Shanghai"

	if s.globalConfig != nil {
		language = s.globalConfig.GetLanguage()
		timeZone = s.globalConfig.GetTimeZone()
	}

	c.JSON(http.StatusOK, gin.H{
		"language": language,
		"timeZone": timeZone,
	})
}

// handleGetSetupStatus 返回系统是否已完成初始化配置（公开接口，无需认证）
// configured=false 表示尚未填写凭据、尚未配置认证方式或尚未创建用户，引导用户前往配置 UI
func (s *Server) handleGetSetupStatus(c *gin.Context) {
	s.mu.RLock()
	cfg := s.config
	authConfigured := len(s.authModes) > 0
	userStore := s.userStore
	s.mu.RUnlock()

	credConfigured := cfg != nil && cfg.AccessKeyId != ""

	// 检查是否存在至少一个用户账号
	usersConfigured := false
	if userStore != nil {
		if users, err := userStore.ListUsers(); err == nil && len(users) > 0 {
			usersConfigured = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"configured":      credConfigured && authConfigured && usersConfigured,
		"credConfigured":  credConfigured,
		"authConfigured":  authConfigured,
		"usersConfigured": usersConfigured,
	})
}
