> #### 惟愿公平如大水滚滚，使公义如江河滔滔。 (阿摩司书 5:24 和合本)

> 作为goagent, shadowsocks, v2ray项目的后继者，记念他们在对抗信息审查所作的努力！


## Backend

```
cd logv2rayfullstack

# 运行程序
go run . server 
# or 
go run ./ s
```

## Frontend

```
cd logv2rayfullstack
npx create-react-app frontend --template redux
cd frontend

npm start

# 生成生产环境文件
npm run build
```

### Thanks To:

> Frontend

- [Protected Routes and Authentication with React Router](https://ui.dev/react-router-protected-routes-authentication/)

> Backend

- [Build user authentication in Golang with JWT and mongoDB](https://dev.to/joojodontoh/build-user-authentication-in-golang-with-jwt-and-mongodb-2igd)