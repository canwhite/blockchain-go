# 区块链项目结构树状图

```
blockchain-go/
├── .env                          # 根目录环境配置，设置HTTP服务端口(8080)
├── go.mod                        # Go模块依赖管理文件
├── go.sum                        # Go模块校验和文件
├── main.go                       # HTTP服务入口，提供REST API接口
├── README.md                     # 项目说明文档
├── networking/                   # TCP网络通信模块
│   ├── .env                      # TCP服务端口配置(9000)
│   └── main.go                   # TCP网络通信实现
├── p2p/                          # P2P网络通信模块
│   ├── .env                      # P2P服务端口配置(9000)
│   ├── main.go                   # P2P网络通信核心实现
│   └── README.md                 # P2P模块说明
├── proof-stake/                  # 权益证明共识模块
│   ├── .env                      # 环境配置
│   └── main.go                   # 权益证明实现
└── proof-work/                   # 工作量证明共识模块
    ├── .env                      # 环境配置
    └── main.go                   # 工作量证明实现
```

# P2P 模块内部结构

```
p2p/main.go
├── 数据结构定义
│   └── Block                     # 区块结构体
├── 全局变量
│   ├── Blockchain[]             # 区块链数组
│   └── mutex                    # 互斥锁
├── 核心函数
│   ├── 区块链操作
│   │   ├── isBlockValid()       # 验证区块有效性
│   │   ├── calculateHash()      # 计算区块哈希
│   │   └── generateBlock()      # 生成新区块
│   ├── P2P网络
│   │   ├── makeBasicHost()      # 创建P2P主机
│   │   ├── handleStream()       # 处理网络流
│   │   ├── readData()           # 读取网络数据
│   │   └── writeData()          # 写入网络数据
│   └── main()                   # 程序入口
└── 命令行参数
    ├── -l                       # 监听端口
    ├── -d                       # 目标节点地址
    ├── -secio                   # 启用安全通信
    └── -seed                    # 随机种子
```
