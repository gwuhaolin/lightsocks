package lightsocks

import (
	"encoding/base64"
	"errors"
	"math/rand"
	"strings"
	"time"
)

const passwordLength = 256

type password [passwordLength]byte

func init() {
	// 更新随机种子，防止生成一样的随机密码
	rand.Seed(time.Now().Unix())
}

// 采用base64编码把密码转换为字符串
func (password *password) String() string {
	return base64.StdEncoding.EncodeToString(password[:])
}

// 解析采用base64编码的字符串获取密码
func parsePassword(passwordString string) (*password, error) {
	bs, err := base64.StdEncoding.DecodeString(strings.TrimSpace(passwordString))
	if err != nil || len(bs) != passwordLength {
		return nil, errors.New("不合法的密码")
	}
	password := password{}
	copy(password[:], bs)
	bs = nil
	return &password, nil
}

// 产生 256个byte随机组合的 密码，最后会使用base64编码为字符串存储在配置文件中
// 不能出现任何一个重复的byte位，必须又 0-255 组成，并且都需要包含
func RandPassword() string {
	password := &password{}
	for i := byte(0); ; i++ {
		password[i] = i
		if i == 255 {
			break
		}
	}
	rand.Shuffle(passwordLength, func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})
	return password.String()
}
