package server

import (
	"encoding/binary"
	"net"
	"time"

	"../core"
)

type LsServer struct {
	*core.SecureSocket
	AfterListen func(listenAddr net.Addr)
}

// 新建一个服务端
// 服务端的职责是:
// 0.监听来自本地代理客户端的请求
// 1.解密本地代理客户端请求的数据，解析socks5协议，连接用户浏览器真正想要连接的远程服务器
// 2.加密后转发用户浏览器真正想要连接的远程服务器返回的数据到本地代理客户端
func New(password *core.Password, listenAddr *net.TCPAddr) *LsServer {
	return &LsServer{
		SecureSocket: &core.SecureSocket{
			Cipher:     core.NewCipher(password),
			ListenAddr: listenAddr,
		},
	}
}

// 运行服务端并且监听来自本地代理客户端的请求
func (server *LsServer) Listen() error {
	listener, err := net.ListenTCP("tcp", server.ListenAddr)
	if err != nil {
		return err
	}

	defer listener.Close()

	if server.AfterListen != nil {
		server.AfterListen(listener.Addr())
	}

	for {
		localConn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		// localConn被关闭时直接清除所有数据 不管没有发送的数据
		localConn.SetLinger(0)
		go server.handleConn(localConn)
	}
	return nil
}

// socks5实现
// https://www.ietf.org/rfc/rfc1928.txt
// http://www.jianshu.com/p/172810a70fad
func (server *LsServer) handleConn(localConn *net.TCPConn) {
	defer localConn.Close()
	buf := make([]byte, 256)

	/**
		The localConn connects to the dstServer, and sends a ver
	   	identifier/method selection message:
			   +----+----------+----------+
	                   |VER | NMETHODS | METHODS  |
	                   +----+----------+----------+
	                   | 1  |    1     | 1 to 255 |
	                   +----+----------+----------+
		The VER field is set to X'05' for this ver of the protocol.  The
	   	NMETHODS field contains the number of method identifier octets that
	   	appear in the METHODS field.
	*/
	// 第一个字段VER代表Socks的版本，Socks5默认为0x05，其固定长度为1个字节
	_, err := server.DecodeRead(localConn, buf)
	// 只支持版本5
	if err != nil || buf[0] != 0x05 {
		return
	}

	/**
		The dstServer selects from one of the methods given in METHODS, and
	   	sends a METHOD selection message:

	                         +----+--------+
	                         |VER | METHOD |
	                         +----+--------+
	                         | 1  |   1    |
	                         +----+--------+
	*/
	// 不需要验证，直接验证通过
	server.EncodeWrite(localConn, []byte{0x05, 0x00})

	/**
		+----+-----+-------+------+----------+----------+
	        |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	        +----+-----+-------+------+----------+----------+
	        | 1  |  1  | X'00' |  1   | Variable |    2     |
	        +----+-----+-------+------+----------+----------+
	*/

	// CMD代表客户端请求的类型，值长度也是1个字节，有三种类型
	// CONNECT X'01'
	if buf[1] != 0x01 {
		// 目前只支持 CONNECT
		return
	}

	n, err := server.DecodeRead(localConn, buf)
	// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
	if err != nil || n < 7 {
		return
	}
	var dIP []byte
	// aType 代表请求的远程服务器地址类型，值长度1个字节，有三种类型
	switch buf[3] {
	case 0x01:
		//	IP V4 address: X'01'
		dIP = buf[4 : 4+net.IPv4len]
	case 0x03:
		//	DOMAINNAME: X'03'
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return
		}
		dIP = ipAddr.IP
	case 0x04:
		//	IP V6 address: X'04'
		dIP = buf[4 : 4+net.IPv6len]
	default:
		return
	}
	dPort := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)

	/**
		 +----+-----+-------+------+----------+----------+
	        |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	        +----+-----+-------+------+----------+----------+
	        | 1  |  1  | X'00' |  1   | Variable |    2     |
	        +----+-----+-------+------+----------+----------+
	*/
	if err != nil {
		return
	} else {
		defer dstServer.Close()
		// 响应客户端连接成功
		server.EncodeWrite(localConn, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		dstServer.SetLinger(0)
		dstServer.SetDeadline(time.Now().Add(core.TIMEOUT))
	}
	// 进行转发
	go server.DecodeCopy(dstServer, localConn)
	server.EncodeCopy(localConn, dstServer)
}
