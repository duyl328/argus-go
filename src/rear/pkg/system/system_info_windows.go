//go:build windows

package system

import (
	"os"
	"os/user"
)

// isWindowsAdmin 检查是否为 Windows 管理员 (简化版本)
func isWindowsAdmin() bool {
	// 简化实现：检查用户名或环境变量
	// 实际应用中应该使用 Windows API

	// 检查是否为 Administrator 用户
	currentUser, err := user.Current()
	if err != nil {
		return false
	}

	// 简单检查用户组
	groups, err := currentUser.GroupIds()
	if err != nil {
		return false
	}

	// 检查是否在管理员组中 (S-1-5-32-544 是内置管理员组的 SID)
	for _, groupId := range groups {
		if groupId == "S-1-5-32-544" {
			return true
		}
	}

	return false
}

// GetWindowsVersion 获取 Windows 版本信息 (简化版本)
func GetWindowsVersion() (string, error) {
	// 简化实现：通过环境变量或注册表获取
	// 实际应用中应该使用 Windows API

	// 简单的版本检测，基于操作系统环境
	if osName := os.Getenv("OS"); osName != "" {
		return "Windows", nil
	}

	return "Windows (Unknown Version)", nil
}

// GetWindowsUserInfo 获取 Windows 用户详细信息
func GetWindowsUserInfo() (*WindowsUserInfo, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	userInfo := &WindowsUserInfo{
		Username:    currentUser.Username,
		Name:        currentUser.Name,
		HomeDir:     currentUser.HomeDir,
		Groups:      []string{},
		IsAdmin:     isWindowsAdmin(),
		UserProfile: os.Getenv("USERPROFILE"),
	}

	// 获取用户组信息
	groups, err := currentUser.GroupIds()
	if err == nil {
		for _, groupId := range groups {
			if group, err := user.LookupGroupId(groupId); err == nil {
				userInfo.Groups = append(userInfo.Groups, group.Name)
			}
		}
	}

	return userInfo, nil
}

// WindowsUserInfo Windows 用户信息
type WindowsUserInfo struct {
	Username    string   `json:"username"`
	Name        string   `json:"name"`
	HomeDir     string   `json:"home_dir"`
	UserProfile string   `json:"user_profile"`
	Groups      []string `json:"groups"`
	IsAdmin     bool     `json:"is_admin"`
}

// GetWindowsSystemInfo 获取 Windows 特有的系统信息
func GetWindowsSystemInfo() (*WindowsSystemInfo, error) {
	info := &WindowsSystemInfo{}

	// 获取版本信息
	if version, err := GetWindowsVersion(); err == nil {
		info.Version = version
	}

	// 获取用户信息
	if userInfo, err := GetWindowsUserInfo(); err == nil {
		info.UserInfo = userInfo
	}

	// 获取系统目录
	info.SystemDir = os.Getenv("SystemRoot")
	info.ProgramFiles = os.Getenv("ProgramFiles")
	info.ProgramFilesX86 = os.Getenv("ProgramFiles(x86)")

	return info, nil
}

// WindowsSystemInfo Windows 系统信息
type WindowsSystemInfo struct {
	Version         string           `json:"version"`
	UserInfo        *WindowsUserInfo `json:"user_info"`
	SystemDir       string           `json:"system_dir"`
	ProgramFiles    string           `json:"program_files"`
	ProgramFilesX86 string           `json:"program_files_x86"`
}

// isUnixRoot Windows 上的占位符函数
func isUnixRoot() bool {
	return false // Windows 上不适用
}