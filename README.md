[TOC]

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

#### namespace 启动
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

#### cgroup 资源限制

对于 cgroup v1 版本：
1. 查看 subsystem 的挂载目录
    - 通过 /proc/self/mountinfo 可以查看当前进程的挂在信息。
    - 默认是将所有的 subsystem 挂载到当前的系统。
```shell
....
[root@compute1 ~]# cat /proc/self/mountinfo
29 22 0:25 / /sys/fs/cgroup ro,nosuid,nodev,noexec shared:4 - tmpfs tmpfs ro,mode=755
30 29 0:26 / /sys/fs/cgroup/systemd rw,nosuid,nodev,noexec,relatime shared:5 - cgroup cgroup rw,xattr,release_agent=/usr/lib/systemd/systemd-cgroups-agent,name=systemd
31 22 0:27 / /sys/fs/pstore rw,nosuid,nodev,noexec,relatime shared:17 - pstore pstore rw
32 22 0:28 / /sys/fs/bpf rw,nosuid,nodev,noexec,relatime shared:18 - bpf bpf rw,mode=700
33 29 0:29 / /sys/fs/cgroup/rdma rw,nosuid,nodev,noexec,relatime shared:6 - cgroup cgroup rw,rdma
34 29 0:30 / /sys/fs/cgroup/cpu,cpuacct rw,nosuid,nodev,noexec,relatime shared:7 - cgroup cgroup rw,cpu,cpuacct
35 29 0:31 / /sys/fs/cgroup/blkio rw,nosuid,nodev,noexec,relatime shared:8 - cgroup cgroup rw,blkio
36 29 0:32 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:9 - cgroup cgroup rw,memory
37 29 0:33 / /sys/fs/cgroup/perf_event rw,nosuid,nodev,noexec,relatime shared:10 - cgroup cgroup rw,perf_event
38 29 0:34 / /sys/fs/cgroup/freezer rw,nosuid,nodev,noexec,relatime shared:11 - cgroup cgroup rw,freezer
39 29 0:35 / /sys/fs/cgroup/cpuset rw,nosuid,nodev,noexec,relatime shared:12 - cgroup cgroup rw,cpuset
40 29 0:36 / /sys/fs/cgroup/net_cls,net_prio rw,nosuid,nodev,noexec,relatime shared:13 - cgroup cgroup rw,net_cls,net_prio
41 29 0:37 / /sys/fs/cgroup/hugetlb rw,nosuid,nodev,noexec,relatime shared:14 - cgroup cgroup rw,hugetlb
42 29 0:38 / /sys/fs/cgroup/devices rw,nosuid,nodev,noexec,relatime shared:15 - cgroup cgroup rw,devices
43 29 0:39 / /sys/fs/cgroup/pids rw,nosuid,nodev,noexec,relatime shared:16 - cgroup cgroup rw,pids
...
```
2. 详细信息 
   - 通过最后的option是rw,memory，可以看出这一条挂载的 subsystem 是 memory，
   - 在 /sys/fs/cgroup/memory 中创建文件夹对应创建的 cgroup，就可以用来做内存的限制
   - 因此通过 mountinfo 文件就可以找到 具体某个 cgroup 的挂载目录
```shell
  36 29 0:32 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:9 - cgroup cgroup rw,memory
```

对于 cgroup v2 版本：
1. 查看 当前进程的 cgroup 信息，在目录 `/proc/self/cgroup`
  - 可以看到，关在的目录就在 /user.slice/user-0.slice/session-9.scope
  - 再加上前缀 /sys/fs/cgroup/user.slice/user-0.slice/session-9.scope
```shell
root@ubuntu:~# cat /proc/self/cgroup
0::/user.slice/user-0.slice/session-9.scope
```
2. docker 在 cgroup v2 时，存放的 cgroup path 为 /sys/fs/cgroup/systemd.slice/docker-xxxx
   - 我们采用在当前用户目录的方式，但是不到 session 的层面，使用 /sys/fs/cgroup/user.slice/user-0.slice
3. 配置：
   - `cpu.max`：文件支持2个值，格式为：`$MAX $PERIOD`。如下就表示：在 100000 所表示的时间周期内，有 50000 是分给本 cgroup 的。也就是配置了本 cgroup 的 cpu 占用在单核上不超过 50%
   ```shell
   [root@localhost]# cat /sys/fs/cgroup/zorro/cpu.max
   50000 100000
   ```

通过 CgroupManager，将资源限制的配置，以及将进程移动到cgroup中的操作交给各个 subsystem 去处理。流程图如下：

![cgroup的流程图](https://github.com/Nevermore12321/LeetCode/blob/blog/%E4%BA%91%E8%AE%A1%E7%AE%97/docker/Cgroup%E7%9A%84%E5%88%9B%E5%BB%BA%E4%B8%8E%E4%BF%AE%E6%94%B9.PNG?raw=true)


#### 构造镜像

**使用 pivot_root 系统调用：**
- pivot_root 是一个系统调用，主要功能是去改变当前的root文件系统。
- pivot_root 可以将当前进程的 root 文件系统移动到 put_old 文件夹中，然后使 new_root 成为新的 root 文件系统。
  - new_root 和 put_old 必须不能同时存在当前 root 的同一个文件系统中。
- pivot_root 和 chroot 的主要区别是:
  - pivot_root 是把整个系统切换到一个新的 root 目录，而移除对之前 root 文件系统的依赖，这样你就能够 umount 原先的 root 文件系统。
  - chroot 是针对某个进程，系统的其他部分依旧运行于老的 root 目录中。

pivot_root: 改变root文件系统
    - pivot_root把当前进程的root文件系统放到put_old目录，而使new_root成为新的root文件系统。
```shell
pivot_root(new_root,put_old)
```

#### 使用 overlay2 来构造镜像

使用 overlay2 来构造镜像，过程：
- Docker在使用镜像启动一个容器时，会新建 2 个 layer
  - writelayer: 是容器唯一的可读写层
  - container-initlayer : 是为容器新建的只读层，用来存储容器启动时传入的系统信息 
- 但是并不是以writelayer和container-initlayer命名的
- 最后把 writelayer、container-initlayer 和相关镜像的 layers 都 mount 到一个 mnt 目录下，然后把这个 mnt 目录作为容器启动的根目录。

根据 overlay2 有三层架构：
- CreateRootDir(rootURL) : 创建存放 dockergsh 层级的 根目录
- CreateLowerLayer(imageURL, rootURL) : 创建 rootURL/lower lower layer 层级目录
- CreateUpperLayer(rootURL) : 创建 rootURL/upper upper layer 层级目录
- CreateWorkDir(rootURL) : 创建 rootURL/work work layer 层级目录
- CreateMountPoint(imageURL, rootURl) : 创建 rootURL/merge merge layer 层级目录

![构造镜像的文件系统层级流程](https://github.com/Nevermore12321/LeetCode/blob/blog/%E4%BA%91%E8%AE%A1%E7%AE%97/docker/%E4%BD%BF%E7%94%A8overlay%E5%88%9B%E5%BB%BA%E6%96%87%E4%BB%B6%E7%B3%BB%E7%BB%9F%E5%B1%82%E7%BA%A7%E8%BF%87%E7%A8%8B.PNG?raw=true)


#### 挂载 volume

挂载流程：
![mount volume](https://github.com/Nevermore12321/LeetCode/blob/blog/%E4%BA%91%E8%AE%A1%E7%AE%97/docker/%E6%8C%82%E8%BD%BDvolume%E6%B5%81%E7%A8%8B.PNG?raw=true)