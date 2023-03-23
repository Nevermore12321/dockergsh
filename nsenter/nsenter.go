package nsenter

/*
#define _GNU_SOURCE
#include <sched.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>


//这里的__attribute__((constructor))指的是， 一旦这个包被引用，那么这个函数就会被自动执行
//类似于构造函数，会在程序一启动的时候运行
__attribute__((constructor)) void enter_namespace(void) {
	char *dockergsh_pid;
	// 从环境变量中读取 容器的 pid
	dockergsh_pid = getenv("dockergsh_pid");
	if (dockergsh_pid) {
		fprintf(stdout, "got dockergsh_pid=%s\n", dockergsh_pid);
	} else {
		fprintf(stdout, "missing dockergsh_pid env skip nsenter\n");
		return;
	}

	char *dockergsh_cmd;
	// 从环境变量中读取 exec 的命令
	dockergsh_cmd = getenv("dockergsh_cmd");
	if (dockergsh_pid) {
		fprintf(stdout, "got dockergsh_cmd=%s\n", dockergsh_cmd);
	} else {
		fprintf(stdout, "missing dockergsh_cmd env skip nsenter\n");
		return;
	}

	// 需要设置的 5 中 namespace
	char *namespaces[] = {"ipc", "uts", "net", "pid", "mnt"};
	int i;
	// 暂存不同 namespace 文件路径
	char nspath[1024];
	for (i = 0; i < 5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", dockergsh_pid, namespaces[i]);
		int fd =open(nspath,O_RDONLY);
		// 将当前进程加入到指定的 namespace 中
		if (setns(fd, 0) == -1) {
			fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
	}

	// 将当前的进程加入到 namespace 中后，执行 docker exec 后的命令
	int res = system(dockergsh_cmd);
	exit(0);
	return;
}
*/
import "C"
