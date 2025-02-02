package gitcontrol

import "strings"

// Auth: 인증 정보 (간단히 ID/Token만 관리)
type GitAuth struct {
	UserID string
	Token  string
}

// (1) 특정 파일에서 ID와 Token 읽기
func LoadAuth(gitUserId string, gitToken string) (*GitAuth, error) {
	auth := &GitAuth{
		UserID: strings.TrimSpace(gitUserId),
		Token:  strings.TrimSpace(gitToken),
	}
	return auth, nil
}
