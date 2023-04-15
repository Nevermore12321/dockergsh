package cmdExec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "github.com/Nevermore12321/dockergsh/nsenter"
	"github.com/sirupsen/logrus"
)

const (
	ENV_EXEC_PID = "dockergsh_pid"
	ENV_EXEC_CMD = "dockergsh_cmd"
)

func ExecInContainer(containerArg string, commandArr []string) error {
	// 根据命令行传递的容器名或者容器id 获取要 exec 容器的 pid
	pid, err := GetContainerPidByArg(containerArg)
	if err != nil {
		logrus.Errorf("Get Container %s pid err error %v", containerArg, err)
		return err
	}

	// 将命令 commandArr 以空格分割，然后放入环境变量 ENV_EXEC_CMD 中
	cmdStr := strings.Join(commandArr, " ")
	logrus.Infof("container pid %s", pid)
	logrus.Infof("command %s", cmdStr)

	// 这里是 实现 docker exec 的关键
	// 再次通过执行 /proc/self/exe 文件，来再次执行一次 dockergsh
	// 传入的参数为 exec，也就是再次执行了 dockergsh exec 命令
	// 但这次执行会带入 环境变量 ENV_EXEC_PID 和 ENV_EXEC_CMD
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 传入环境变量，用来控制让 C 代码开始执行
	_ = os.Setenv(ENV_EXEC_CMD, cmdStr)
	_ = os.Setenv(ENV_EXEC_PID, pid)

	// 关键点，每次在 exec 到容器中时，要和容器启动时的环境变量一致
	// 这里调用了 cgo 方法，直接调用linux setns 系统调用，因此继承的是宿主机的环境变量，这一步就是将容器内进程的环境变量加入到 cgo 进程中
	containerEnvs, err := GetEnvsByPid(pid)
	if err != nil {
		logrus.Errorf("Get Envs error %v", err)
		return err
	}
	cmd.Env = append(os.Environ(), containerEnvs...)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Exec container %s error %v", containerArg, err)
		return err
	}
	return nil
}

func GetContainerPidByArg(containerArg string) (string, error) {
	// 获取 containerInfo，
	containerInfo, err := GetContainerInfoByArg(containerArg)
	if err != nil {
		logrus.Errorf("Get Container %s Info err error %v", containerArg, err)
		return "", err
	}

	return containerInfo.Pid, nil
}

// 根据指定的 pid 获取进程的 所有环境变量
// 根据指定的 pid 获取对应进程的环境变量
func GetEnvsByPid(pid string) ([]string, error) {
	// 进程的环境变量信息存放在 /proc/[PID]/environ 文件中
	// 获取对应 pid 的环境变量文件路径
	path := fmt.Sprintf("/proc/%s/environ", pid)

	// 读取环境变量文件内容
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		logrus.Errorf("Read file %s error %v", path, err)
		return nil, err
	}

	// 多个环境变量之间的分隔符是 \u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs, nil
}
