package mysocks

import (
	"net"
	"encoding/binary"
)

// LsServer 定义Server端
type LsServer struct {
	Cipher *cipher
	ListenAddr *net.TCPAddr
}

// NewListenServer 新建一个服务端，服务端的职责：
// 1、监听来自本地客户端的请求
// 2、解密本地客户端请求的数据，解析 socks5 协议，连接用户真正想要连接的远程服务器
// 3、转发用户想要真正连接的远程服务器返回的数据，加密后发到本地客户端
func NewListenServer(password string, listenAddr string) (*LsServer, error) {
	bsPassword, err := parsePassword(password)
	if err != nil {
		return nil, err
	}
	structListenAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &LsServer{
		Cipher: newCipher(bsPassword),
		ListenAddr: structListenAddr,
	}, nil
}

// Listen 运行服务端并监听来自本地客户端的请求
func (lsServer *LsServer) Listen(didListen func(listenAddr net.Addr)) error {
	return ListenTCPSecure(lsServer.ListenAddr, lsServer.Cipher, lsServer.handleConn, didListen)
}

// handleConn 解析 Socks5 协议
func (lsServer *LsServer) handleConn(localSecureConn *SecureTCPConn) {
	defer localSecureConn.Close()
	buf := make([]byte, 256)

	_, err := localSecureConn.DecodeRead(buf)
	if err != nil || buf[0] != 0x05 {
		return 
	}

	localSecureConn.EncodeWrite([]byte{0x05, 0x00})  // 不需要验证，直接通过

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/
	n, err := localSecureConn.DecodeRead(buf)
	if err != nil || n < 7 {  // 最短为 7, addressType=0x03(域名类型)且dst.addr占1个字节
		return 
	}
	if buf[1] != 0x01 {  // 只支持 0x01 connect类型
		return 
	}
	var dIP []byte
	switch buf[3] {
	case 0x01:
		dIP = buf[4:4+net.IPv4len]
	case 0x03:
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return 
		}
		dIP =  ipAddr.IP
	case 0x04:
		dIP = buf[4:4+net.IPv6len]
	default:
		return 
	}
	dPort := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP: dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),  // 网络序使用大端序传输
	}
	// 和真正的远端服务器建立连接
	trueSecureConn, err := net.DialTCP("tcp", nil, dstAddr) 
	if err != nil {
		return 
	}
	defer trueSecureConn.Close()
	trueSecureConn.SetLinger(0)
	// 和客户端说，真正的服务器连接成功
	localSecureConn.EncodeWrite([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	
	// 进行转发,从localSecureConn获取数据，解密发送到 trueSecureConn
	go func() {
		err := localSecureConn.DecodeCopy(trueSecureConn)
		if err != nil {
			localSecureConn.Close()
			trueSecureConn.Close()
		}
	} ()
	(&SecureTCPConn{
		Cipher: localSecureConn.Cipher,
		ReadWriteCloser: trueSecureConn,
	}).EncodeCopy(localSecureConn)
}
