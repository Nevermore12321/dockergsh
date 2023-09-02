package registry

// 在 engine 对象对外暴露的API信息中添加 dockergsh registry 的信息。
// 若 registry.NewService（）被成功安装，则会有两个相应的处理方法注册至 engine:
// 1. Dockergsh Daemon 通过 Docker Client 提供的认证信息向 registry 发起认证请求；
// 2. search，在公有registry上搜索指定的镜像，目前公有的registry只支持 Docker Hub。
