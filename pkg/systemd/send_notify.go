package systemd

import (
	"errors"
	"net"
	"os"
)

var SendNotifyNoSocket = errors.New("no socket found")

// 发送消息到 init daemon
func SendNotify(state string) error {
	// `Net: "unixgram"` 指定了使用 “数据报”（datagram）Unix 套接字类型。
	//-如果服务被 `systemd` 托管运行， 会由 `systemd` 设置，用于指定通信的套接字地址。 `NOTIFY_SOCKET`
	socketAddr := &net.UnixAddr{
		Name: os.Getenv("NOTIFY_SOCKET"),
		Net:  "unixgram",
	}

	if socketAddr.Name == "" {
		return SendNotifyNoSocket
	}

	// 连接到 NOTIFY_SOCKET 提供的 Unix 套接字地址。
	conn, err := net.DialUnix("unixgram", nil, socketAddr)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(state))
	if err != nil {
		return err
	}
	return nil
}
