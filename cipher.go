package mysocks

type cipher struct {
	encodePassword *password
	decodePassword *password
}

// encode 加密原始数据
func(cipher *cipher) encode(bs []byte) {
	for i, v := range bs {
		bs[i] = cipher.encodePassword[v]
	}
}

// decode 解密加密后的数据
func(cipher *cipher) decode(bs []byte) {
	for i, v := range bs {
		bs[i] = cipher.decodePassword[v]
	}
}

// newCipher 新建一个编码/解码器
func newCipher(encodePassword *password) *cipher {
	decodePassword := &password{}
	for i, v := range encodePassword {
		encodePassword[i] = v
		decodePassword[v] = byte(i)
	}
	return &cipher{
		encodePassword: encodePassword,
		decodePassword: decodePassword,
	}
}