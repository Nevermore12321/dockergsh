package utils

var (
	GITCOMMIT string            // GitCommit 信息
	VERSION   string = "v1.0.0" // dockergsh version 信息

	IAMSTATIC bool   // 通过 .hackmake.sh 二进制文件静态编译
	INITSHA1  string // 如果 Dockergsh 本身是通过 .hackmake.sh dynbinary 动态编译的，则单独静态 dockergshinit 的 sha1sum，
	INITPATH  string // 用于搜索有效 dockergshinit 二进制文件的自定义位置（包装商可将其作为最后的逃生手段）
)
