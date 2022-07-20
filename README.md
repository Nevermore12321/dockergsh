# 自己动手写 docker 笔记


## Run 命令

### 基础知识

linux 系统中 /proc 下的几个目录详情：
1. `/proc/N`: PID为N的进程信息进程 
2. `/proc/N/cmdline`: 命令启动命令
3. `/proc/N/cwd`: 链接到进程当前工作目录
4. `/proc/N/environ`: 进程环境变量列表
5. `/proc/N/exe`: 链接到进程的执行命令文件
6. `/proc/N/fd`: 包含进程相关的所有文件描述符
7. `/proc/N/maps`: 与进程相关的内存映射信息
8. `/proc/N/mem`: 指代进程持有的内存，不可读
9. `/proc/N/root`: 链接到进程的根目录
10. `/proc/N/stat`: 进程的状态
11. `/proc/N/statm`: 进程使用的内存状态
12. `/proc/N/status`: 进程状态信息，比stat/statm更具可读性
13. `/proc/self/`: 链接到当前正在运行的进程

### docker run 命令流程

整个 docker run 命令的启动如下：
![docker run 命令流程](https://github.com/Nevermore12321/LeetCode/blob/blog/%E4%BA%91%E8%AE%A1%E7%AE%97/docker/%E8%87%AA%E5%B7%B1%E5%8A%A8%E6%89%8B%E5%86%99docker-run%E5%91%BD%E4%BB%A4%E6%B5%81%E7%A8%8B.PNG?raw=true)

RunContainerInitProcess 的主要作用就是 **使得 docker init 进程为容器内的 1 号进程，也就是容器内的 init 进程**。

本函数最后的 `syscall.Exec()`，是最为重要的一句黑魔法，正是这个系统调用实现了完成初始化动作井将用户进程运行起来的操作。

使用`syscall.Exec()` 原因：
- 使用 Docker 创建起来一个容器之后，会发现容器内的第一个程序，也就是 PID 为 1 的那个进程，是指定的前台进程。
- 容器创建之后，执行的第一个进程并不是用户的进程，而是 init 初始化的进程。这时候，如果通过 ps 命令查看就会发现，容器内第一个进程变成了自己的 init，这和预想的是不一样的

`syscall.Exec()` 这个方法，其实最终调用了 Kernel 的函数：
```
int execve(const char *filename, char *const argv[], char *const envp［］）
```
- 作用是执行当前 filename 对应的程序。
- 它会覆盖当前进程的镜像、数据和堆械等信息，包括 PID，这些都会被将要运行的进程覆盖掉。
- 也就是说，**调用这个方法，将用户指定的进程运行起来，把最初的 init 进程给替换掉，这样当进入到容器内部的时候，就会发现容器内的第一个程序就是我们指定的进程了**。
- 其实也是目前 Docker 使用的容器引擎 runC 的实现方式之一。