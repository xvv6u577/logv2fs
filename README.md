> #### 惟愿公平如大水滚滚，使公义如江河滔滔。 (阿摩司书 5:24 和合本)

> 作为goagent, shadowsocks, v2ray项目的后继者，记念他们在对抗信息审查所作的努力！

## Backend

> 环境变量：.env

```
cd logv2rayfullstack

# 运行程序
go run . server 
# or 
go run ./ s

# 清空数据库
go run ./ mongo delallusers
go run ./ mongo delalltables
```

以systemd service运行（以ubuntu 18.04为例）
```
sudo mv pre-setup/logv2rayfullstack.service /etc/systemd/system/
sudo systemctl daemon-reload

sudo systemctl start logv2rayfullstack.service
sudo systemctl status logv2rayfullstack.service
sudo systemctl stop logv2rayfullstack.service
```

## Frontend

> 环境变量：frontend/.env

```
cd logv2rayfullstack
npx create-react-app frontend --template redux
cd frontend

npm start
```
### 生成生产环境文件
```
npm run build
```
### 辅助脚本 ./tool/ 
#### 生成 users.json
```
node ./tool/generateUsers.js
```
#### 把users.json,db.json数据添加到mongoDB (wild mode有效)
```
node ./tool/pushUsersToDB.js
node ./tool/pushJsonFileToDB.js
```



### Thanks To:

> Frontend

- [Protected Routes and Authentication with React Router](https://ui.dev/react-router-protected-routes-authentication/)

> Backend

- [Build user authentication in Golang with JWT and mongoDB](https://dev.to/joojodontoh/build-user-authentication-in-golang-with-jwt-and-mongodb-2igd)