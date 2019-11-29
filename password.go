package mysocks

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const passwordLength = 256

type password [passwordLength]byte // 定义 password 类型

func init() {
	rand.Seed(time.Now().Unix()) // seed 参数为一个 int64
}

// CreateRandPassword 产生一个 0--255 的错排凯撒密码，返回一个使用 base64 编码的字符串
func CreateRandPassword() string {
	permArr := rand.Perm(passwordLength)
	password := &password{}
	for i, v := range permArr {
		if i == v {
			return CreateRandPassword()
		}
	}
	for i, v := range permArr {
		password[i] = byte(v)
	}
	fmt.Println(password[:])
	return base64.StdEncoding.EncodeToString(password[:]) // password[:] 为 []uint8 类型
}

// parsePassword 解析 base64 编码的密码，得到源密码
func parsePassword(passwordString string) (*password, error) {
	pw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(passwordString))
	if err != nil || len(pw) != passwordLength {
		return nil, errors.New("不合法的密码")
	}
	password := &password{}
	copy(password[:], pw) // pw 的类型为 []byte 切片
	return password, nil
}
