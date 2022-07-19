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

