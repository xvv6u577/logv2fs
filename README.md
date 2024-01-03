> #### 惟愿公平如大水滚滚，使公义如江河滔滔。 (阿摩司书 5:24 和合本)

> 作为goagent, shadowsocks, v2ray项目的后继者，记念他们在对抗信息审查所作的努力！

## Backend

#### gRPC

> Open Ports: 80、443

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/myproto.proto
```

> 环境变量：.env
> go v1.18.10
> v2ray v4.23.4

```
cd logv2fs

# 运行程序
go run . 

go build
go install
```

以systemd service运行（以ubuntu 18.04为例）
```
sudo systemctl daemon-reload

sudo systemctl enable logv2fs.service
sudo systemctl start logv2fs.service

sudo systemctl stop logv2fs.service
sudo systemctl status logv2fs.service
```

## Frontend

> 环境变量：frontend/.env
> nodejs v17.2.0
> npm v8.1.4

### 生成生产环境文件
```
cd logv2fs/frontend
npm i
npm run build
```

### Thanks To:

> Frontend

- [Protected Routes and Authentication with React Router](https://ui.dev/react-router-protected-routes-authentication/)

> Backend

- [Build user authentication in Golang with JWT and mongoDB](https://dev.to/joojodontoh/build-user-authentication-in-golang-with-jwt-and-mongodb-2igd)