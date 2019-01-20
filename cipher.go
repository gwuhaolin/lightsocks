package lightsocks

type cipher struct {
	// 编码用的密码
	encodePassword *password
	// 解码用的密码
	decodePassword *password
}

// 加密原数据
func (cipher *cipher) encode(bs []byte) {
	for i, v := range bs {
		bs[i] = cipher.encodePassword[v]
	}
}

// 解码加密后的数据到原数据
func (cipher *cipher) decode(bs []byte) {
	for i, v := range bs {
		bs[i] = cipher.decodePassword[v]
	}
}

// 新建一个编码解码器
func newCipher(encodePassword *password) *cipher {
	decodePassword := &password{}
	for i, v := range encodePassword {
		decodePassword[v] = byte(i)
	}
	return &cipher{
		encodePassword: encodePassword,
		decodePassword: decodePassword,
	}
}
