package crypto

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/jimstudt/http-authentication/basic"
	"github.com/magic-lib/go-plat-utils/utils"
	"golang.org/x/crypto/bcrypt"
)

type Htpasswd struct {
}

// Apr1Md5Password 生成格式
// 在Windows, Netware 和TPF上，这是默认的加密方式
func (h *Htpasswd) Apr1Md5Password(pwd, salt string) string {
	if salt == "" {
		salt = utils.RandomString(8) //随机生成字符串
	}
	hash := apr1Md5(pwd, salt)
	return fmt.Sprintf("$apr1$%s$%s", salt, hash)
}

// Apr1Md5PasswordCompare 进行判断
func (h *Htpasswd) Apr1Md5PasswordCompare(hash, pwd string) (bool, error) {
	md5Obj, err := basic.AcceptMd5(hash)
	if err != nil {
		return false, err
	}
	return md5Obj.MatchesPassword(pwd), nil
}

// Sha1Password 生成格式
func (h *Htpasswd) Sha1Password(pwd string) string {
	hash := sha1.Sum([]byte(pwd))
	hashStr := base64.StdEncoding.EncodeToString(hash[:])
	return fmt.Sprintf("{SHA}%s", hashStr)
}

// Sha1PasswordCompare 进行判断
func (h *Htpasswd) Sha1PasswordCompare(hash, pwd string) (bool, error) {
	shaObj, err := basic.AcceptSha(hash)
	if err != nil {
		return false, err
	}
	return shaObj.MatchesPassword(pwd), nil
}

// BcryptPassword 加密密码
func (h *Htpasswd) BcryptPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedPassword)
}

// BcryptPasswordCompare 校验密码
func (h *Htpasswd) BcryptPasswordCompare(hash, pwd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
	return err == nil, err
}
