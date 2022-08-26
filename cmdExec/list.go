package cmdExec

import (
	"encoding/json"
	"fmt"
	"github.com/Nevermore12321/dockergsh/container"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/tabwriter"
)

func ListContainers() {
	// /var/lib/dockergsh/
	rootURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	rootURL = rootURL[:len(rootURL) - 1]
	files, err := ioutil.ReadDir(rootURL)
	if err != nil {
		log.Errorf("Read dir %s error %v", rootURL, err)
	}

	// 每个容器的 containerinfo
	var containers []*container.ContainerInfo

	// 遍历每一个容器
	for _, file := range files {
		if file.Name() == "network" || file.Name() == "images" || file.Name() == "containers" {
			continue
		}

		configURL := filepath.Join(rootURL, file.Name(), "container", container.ConfigName)
		// 读取配置文件 config.json
		info, err := getContainerInfo(configURL)
		if err != nil {
			if err != os.ErrInvalid {
				log.Errorf("Get container info error %v", err)
			}
			continue
		}
		containers = append(containers, info)
	}

	// 格式化输出
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		log.Errorf("Format print error: %v", err)
		return
	}

	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreateTime)
	}

	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return
	}
}


/*
读取对应 container 的配置文件 config.json ，并且解析为 ContainerInfo 结构体返回
 */
func getContainerInfo(configURL string) (*container.ContainerInfo, error) {
	// 读取配置文件
	content, err := os.ReadFile(configURL)
	if err != nil {
		log.Errorf("Read file %s error %v", configURL, err)
		return nil, err
	}

	var containerInfo container.ContainerInfo
	// 将配置文件解析为结构体
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("Json unmarshal error %v", err)
		return nil, err
	}

	return &containerInfo, nil
}