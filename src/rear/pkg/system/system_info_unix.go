//go:build !windows

package system

import (
	"os"
	"os/user"
)

// isUnixRoot 检查是否为 Unix root 用户
func isUnixRoot() bool {
	return os.Geteuid() == 0
}

// GetUnixUserInfo 获取 Unix 用户详细信息
func GetUnixUserInfo() (*UnixUserInfo, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	userInfo := &UnixUserInfo{
		Username: currentUser.Username,
		Name:     currentUser.Name,
		HomeDir:  currentUser.HomeDir,
		UID:      currentUser.Uid,
		GID:      currentUser.Gid,
		Groups:   []string{},
		IsRoot:   isUnixRoot(),
		Shell:    os.Getenv("SHELL"),
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

// UnixUserInfo Unix 用户信息
type UnixUserInfo struct {
	Username string   `json:"username"`
	Name     string   `json:"name"`
	HomeDir  string   `json:"home_dir"`
	UID      string   `json:"uid"`
	GID      string   `json:"gid"`
	Groups   []string `json:"groups"`
	IsRoot   bool     `json:"is_root"`
	Shell    string   `json:"shell"`
}

// GetUnixSystemInfo 获取 Unix 特有的系统信息
func GetUnixSystemInfo() (*UnixSystemInfo, error) {
	info := &UnixSystemInfo{}

	// 获取用户信息
	if userInfo, err := GetUnixUserInfo(); err == nil {
		info.UserInfo = userInfo
	}

	// 获取系统路径
	info.RootDir = "/"
	info.UsrDir = "/usr"
	info.VarDir = "/var"
	info.EtcDir = "/etc"
	info.TmpDir = "/tmp"

	return info, nil
}

// UnixSystemInfo Unix 系统信息
type UnixSystemInfo struct {
	UserInfo *UnixUserInfo `json:"user_info"`
	RootDir  string        `json:"root_dir"`
	UsrDir   string        `json:"usr_dir"`
	VarDir   string        `json:"var_dir"`
	EtcDir   string        `json:"etc_dir"`
	TmpDir   string        `json:"tmp_dir"`
}

// isWindowsAdmin Unix 上的占位符函数
func isWindowsAdmin() bool {
	return false // Unix 上不适用
}