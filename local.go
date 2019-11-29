package mysocks

import (
	"net"
	"log"
)

// LsLocal 定义本地端
type LsLocal struct {
	Cipher     *cipher
	ListenAddr *net.TCPAddr
	RemoteAddr *net.TCPAddr
}

// NewListenLocal 新建一个本地端，本地端的职责为：
// 1、监听来自本机浏览器的代理请求（127.0.0.1:80）
// 2、转发前加密数据
// 3、转发socket数据给墙外代理服务器
// 4、把墙外代理服务器返回的数据转发给用户的浏览器
func NewListenLocal(password string, listenAddr, remoteAddr string) (*LsLocal, error) {
	bsPassword, err := parsePassword(password)
	if err != nil {
		return nil, err
	}
	structListenAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	structRemoteAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}
	return &LsLocal{
		Cipher:     newCipher(bsPassword),
		ListenAddr: structListenAddr,
		RemoteAddr: structRemoteAddr,
	}, nil
}

// Listen 本地端启动监听，接收来自本机浏览器的连接
func (local *LsLocal) Listen(didListen func(listenAddr net.Addr)) error {
	return ListenTCPSecure(local.ListenAddr, local.Cipher, local.handleConn, didListen)
}

func (local *LsLocal) handleConn(localSecureConn *SecureTCPConn) {
	defer localSecureConn.Close()
	proxySecureConn, err := DialTCPSecure(local.RemoteAddr, local.Cipher)
	if err != nil {
		log.Println(err)
		return 
	}
	defer proxySecureConn.Close()
	// 从proxySecureConn读取数据发送到localSecureConn
	go func() {
		err := proxySecureConn.DecodeCopy(localSecureConn)
		if err != nil { // decodeCopy 无数据可读，返回 nil，不影响
			localSecureConn.Close()
			proxySecureConn.Close()
		} 
	}()
	// 从localSecureConn发送数据到proxySecureConn
	localSecureConn.EncodeCopy(proxySecureConn)
}
