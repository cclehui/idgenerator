## idGenerator
全局唯一id生成器， 分布式，多存储模式支持

id generator server

## How to use

1.编辑 config/production.toml 文件， 配置自己合适的配置

2.启动master server:  go run server.go master

3.启动slave server: go run server.go slave

3. http://0.0.0.0:8182/autoincrement?source=aaaa


## Contribute
