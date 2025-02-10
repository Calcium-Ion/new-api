package service

import (
	"fmt"
	"one-api/common"
	"one-api/model"

	"github.com/go-ldap/ldap/v3"
)

type LDAPService struct {
	conn *ldap.Conn
}

func NewLDAPService() (*LDAPService, error) {
	if !common.LDAPAuthEnabled {
		return nil, fmt.Errorf("LDAP is not enabled")
	}

	// 连接到 LDAP 服务器
	conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", common.LDAPHost, common.LDAPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %v", err)
	}

	// 使用管理员账号绑定
	err = conn.Bind(common.LDAPBindUsername, common.LDAPBindPassword)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind with admin account: %v", err)
	}

	return &LDAPService{conn: conn}, nil
}

func (s *LDAPService) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *LDAPService) Authenticate(username, password string) (*model.User, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("LDAP connection is not initialized")
	}

	// 搜索用户
	searchRequest := ldap.NewSearchRequest(
		common.LDAPBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(common.LDAPUserFilter, ldap.EscapeFilter(username)),
		[]string{common.LDAPEmailAttr, common.LDAPNameAttr, "dn"},
		nil,
	)

	sr, err := s.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search user: %v", err)
	}

	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("user not found or multiple users found")
	}

	userdn := sr.Entries[0].DN

	// 验证用户密码
	err = s.conn.Bind(userdn, password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 重新绑定管理员账号
	err = s.conn.Bind(common.LDAPBindUsername, common.LDAPBindPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to rebind admin account: %v", err)
	}

	// 获取用户信息
	email := sr.Entries[0].GetAttributeValue(common.LDAPEmailAttr)
	displayName := sr.Entries[0].GetAttributeValue(common.LDAPNameAttr)
	if displayName == "" {
		displayName = username
	}

	// 检查用户是否已存在
	user := &model.User{}
	exist, err := model.CheckUserExistOrDeleted(username, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %v", err)
	}

	if !exist {
		// 创建新用户
		user = &model.User{
			Username:    username,
			DisplayName: displayName,
			Email:       email,
			Role:        1, // 普通用户角色
			Status:      1, // 启用状态
		}
		if err := user.Insert(0); err != nil {
			return nil, fmt.Errorf("failed to create user: %v", err)
		}
	} else {
		// 获取已存在的用户
		user.Username = username
		if err := user.FillUserByUsername(); err != nil {
			return nil, fmt.Errorf("failed to get user: %v", err)
		}
		// 更新用户信息
		user.DisplayName = displayName
		user.Email = email
		if err := user.Update(false); err != nil {
			return nil, fmt.Errorf("failed to update user: %v", err)
		}
	}

	return user, nil
}
