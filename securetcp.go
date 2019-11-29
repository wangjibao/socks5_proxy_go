package mysocks

import (
	"io"
	"log"
	"net"
)

const bufSize = 1024

// SecureTCPConn 加密传输的 TCP 连接
type SecureTCPConn struct {
	io.ReadWriteCloser // 整合了 Read、Write 和 Closer 的接口, 任何实现了这三个函数的类型都可传递给该接口类型（多态）
	Cipher             *cipher
}

// DecodeRead 从securesocket中读取加密的数据并解密，原数据放到 bs 里
func (secureSocket *SecureTCPConn) DecodeRead(bs []byte) (n int, err error) {
	n, err = secureSocket.Read(bs)
	if err != nil {
		return // 由于上面使用了"命名返回值"，所以此处直接使用简单的 return 即可
	}
	secureSocket.Cipher.decode(bs[:n])
	return
}

// EncodeWrite 把bs中的数据加密后，写入到 securesocket 中
func (secureSocket *SecureTCPConn) EncodeWrite(bs []byte) (int, error) {
	secureSocket.Cipher.encode(bs)
	return secureSocket.Write(bs)
}

// EncodeCopy 从 secureSocket 中不断读取源数据，加密写入到 dst 中，直到 secureSocket 中无数据可读
func (secureSocket *SecureTCPConn) EncodeCopy(dst io.ReadWriteCloser) error {
	buf := make([]byte, bufSize) // buf 切片，大小和容量均为 bufSize
	for {
		readCount, errRead := secureSocket.Read(buf)
		if errRead != nil {
			if errRead == io.EOF {
				return nil
			}
			return errRead
		}
		if readCount > 0 {
			writeCount, errWrite := (&SecureTCPConn{
				ReadWriteCloser: dst,
				Cipher:          secureSocket.Cipher,
			}).EncodeWrite(buf[0:readCount]) // 调用的是 dst 实现的 write
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}

// DecodeCopy 从 secureSocket 中不断读取加密的数据，解密后写入到 dst 中，直到 secureSocket 中无数据可读
func (secureSocket *SecureTCPConn) DecodeCopy(dst io.Writer) error {
	buf := make([]byte, bufSize)
	// 读取完成退出 for 循环
	for {
		readCount, errRead := secureSocket.DecodeRead(buf)
		if errRead != nil {
			if errRead == io.EOF {
				return nil
			}
			return errRead
		}
		if readCount > 0 {
			writeCount, errWrite := dst.Write(buf[0:readCount])
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}

// DialTCPSecure 实现的功能和 net.dialTCP 是一样的，只不过加了一个安全的功能
func DialTCPSecure(remoteAddr *net.TCPAddr, cipher *cipher) (*SecureTCPConn, error) {
	remoteConn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}
	return &SecureTCPConn{
		ReadWriteCloser: remoteConn,
		Cipher:          cipher,
	}, nil
}

// ListenTCPSecure 实现的功能和 ListenTCP 是一样的，加了一个安全的功能
func ListenTCPSecure(localAddr *net.TCPAddr, cipher *cipher, handleConn func(secureConn *SecureTCPConn), didListen func(listenAddr net.Addr)) error {
	listener, err := net.ListenTCP("tcp", localAddr) // 使用 localAddr 创建一个监听器
	if err != nil {
		return err
	}
	defer listener.Close()
	if didListen != nil {
		didListen(listener.Addr())
	}
	for {
		localTCPConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		// localConn 被关闭时，直接清除所有数据
		localTCPConn.SetLinger(0)
		go handleConn(&SecureTCPConn{
			ReadWriteCloser: localTCPConn,
			Cipher: cipher,
		})
	}
}
