package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
)

func PreCheckConfTLS(context *cli.Context) (*tls.Config, error) {
	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = true
	// tlsConfig 对象需要加载一个受信的 ca 文件
	// 如果flTlsVerify为true，Docker Client连接Docker Server需要验证安全性
	if context.Bool("tls-verify") {
		// 读取 ca 证书
		filePath := context.String("tls-cacert")
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("Couldn't read ca cert %s: %s", filePath, err)
		}

		// 证书池是用于存储根证书和中间证书的集合，用于在验证证书链时提供验证所需的信任锚点。
		certPool := x509.NewCertPool()

		// 向证书池添加根证书
		certPool.AppendCertsFromPEM(file)

		tlsConfig.RootCAs = certPool
		tlsConfig.InsecureSkipVerify = false

		// 如果flTlsVerify有一个为真，那么需要加载证书发送给客户端。
		_, errCert := os.Stat(context.String("tls-cert"))
		_, errKey := os.Stat(context.String("tls-key"))
		if errCert == nil && errKey == nil {
			cert, err := tls.LoadX509KeyPair(context.String("tls-cert"), context.String("tls-key"))
			if err != nil {
				return nil, fmt.Errorf("Couldn't load X509 key pair: %s. Key encrypted?", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		return &tlsConfig, nil
	} else {
		return nil, nil
	}
}
