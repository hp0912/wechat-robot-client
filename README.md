# 机器人客户端

## 免责声明

**本项目仅供学习交流使用，严禁用于商业用途**

使用本项目所产生的一切法律责任和风险，由使用者自行承担，与项目作者无关。

请遵守相关法律法规，合法合规使用本项目。

## 特别声明

**本项目仅部分源代码公开**

- 机器人管理后台

  - 前端项目: [https://github.com/hp0912/wechat-robot-admin-frontend](https://github.com/hp0912/wechat-robot-admin-frontend)

  - 后端项目: [https://github.com/hp0912/wechat-robot-admin-backend](https://github.com/hp0912/wechat-robot-admin-backend)

- 机器人客户端和服务端

  - 机器人客户端: [本项目](https://github.com/hp0912/wechat-robot-client)

  - 机器人服务端 **(源代码不公开)** [接口文档](ipad.swagger.yml)

- 公共服务

  - 公众号认证服务: [https://github.com/hp0912/wechat-server](https://github.com/hp0912/wechat-server) fork的项目，微信公众号的后端，为管理后台(以及其他系统)提供微信登录验证功能

  - 词云服务: [https://github.com/hp0912/word-cloud-server](https://github.com/hp0912/word-cloud-server) golang写的词云效果不太好，用python写了一个单独的服务

> 服务端采用iPad协议，可以去马老板开的动物园淘一淘

## 项目概览

本项目是一个智能机器人管理系统，提供了丰富的交互体验。

- AI聊天，chat-gtp deepseek qwen 系列等等

- AI绘图，豆包文生图，智谱文生图，即梦文生图，豆包图像编辑

- AI语音，文本转语音，长文本转语音

- 群聊欢迎新成员，支持文本、图片、表情包、链接形式

- 点歌

- 群聊退群提醒

- 拍一拍交互

- 群聊每日、每周、每月活跃排行榜，每日群聊词云

- 抖音短链接视频解析

- 群聊每日总结

- 群聊每日早报

- 收藏夹 (待开发)

- 查看朋友圈，查看指定人的朋友圈，自动评论、点赞

- 手动添加好友(搜索添加、从群聊添加)

- 手动通过好友验证，自动通过好友验证

- 手动同意进入群聊，自动拉人进入群聊

- 手动发起群聊

- 授权登录APP(王者荣耀、吃鸡等等)

- 配合【推送加】，支持掉线通知，推送到指定微信

## 使用方式

**自部署**

> 自部署前的准备
>
> - 你得有自己的公众号，用来集成公众号扫码登录，本项目只集成了公众号扫码登录
>
> - 自己会安装 docker 和 docker-compose

**直接使用现成系统**

> 访问 [https://wechat-sz.houhoukang.com/](https://wechat-sz.houhoukang.com/) 使用微信扫码登录管理后台，进入后台后创建微信机器人实例。使用微信扫码登录机器人（iPad）。
>
> 风险提示：本机器人服务器在广东，非广东地区的慎重使用，微信异地登陆有概率被风控。

### 自部署基础篇

#### 启动服务

```vim
# 克隆本项目
git clone git@github.com:hp0912/wechat-robot-client.git

# 进入部署目录
cd ./wechat-robot-client/.deploy/local

# 通过docker-compose启动容器，下面两个命令，哪个能用就用哪个
docker compose up -d
docker-compose up -d
```

#### 配置公众号认证服务

1. 访问 http://127.0.0.1:8090 **微信服务器**

2. 配置 **微信服务器**

> 如何配置，前往 [https://github.com/hp0912/wechat-server](https://github.com/hp0912/wechat-server) 查看详细教程。
> 
> 在**微信服务器** `设置` `个人设置` `生成访问令牌`生成的令牌，填入`docker-compose.yml`的`WECHAT_SERVER_TOKEN`的环境变量中，将你自己的公众号二维码链接填入`WECHAT_OFFICIAL_ACCOUNT_AUTH_URL`环境变量中。

3. 重启服务

```
docker compose up -d
docker-compose up -d
```

4. 访问 http://127.0.0.1:8080 **机器人管理后台**

5. 使用个人微信扫码登录

6. 新建机器人

### 自部署进阶篇

**部署在远程服务器**

> 自部署前的准备 (跟本地部署一样，只不过服务器安装docker有点门槛)
>
> - 你得有自己的公众号，用来集成公众号扫码登录，本项目只集成了公众号扫码登录
>
> - 服务器安装 docker 和 docker-compose
>
> - 服务器安装 nginx
>
> - 域名解析，将你的自定义域名解析到你自己的服务器(有公网IP)

```vim
# 克隆本项目
git clone git@github.com:hp0912/wechat-robot-client.git

# 进入部署目录
cd ./wechat-robot-client/.deploy/server

# 通过docker-compose启动容器，下面两个命令，哪个能用就用哪个
docker compose up -d
docker-compose up -d
```

**修改nginx配置**

> `.deploy/server/nginx.conf`这个文件配置了服务转发规则，`wechat-server.xxx.com`(改成你自己的域名) 域名转发到3000端口，也就是`docker-compose.yml`里面的`wechat-server`服务。
>
>`wechat-robot.xxx.com`(改成你自己的域名) 域名，`api/v1`开头的路由转发到3002端口，也就是`docker-compose.yml`里面的`wechat-robot-admin-backend`服务，剩下路由转发到3001端口，也就是`docker-compose.yml`里面的`wechat-robot-admin-fontend`服务
>
> 将这个文件覆盖服务器上的 `/etc/nginx/sites-available/default`

**重启nginx服务**

```
sudo service nginx restart
```

**使用 Let's Encrypt 的 certbot 配置 HTTPS**

> 需要先配置好域名解析

```
# Ubuntu 安装 certbot：
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
# 生成证书 & 修改 Nginx 配置
sudo certbot --nginx
# 根据指示进行操作
# 重启 Nginx
sudo service nginx restart
```

**配置微信服务器，获取`WECHAT_SERVER_TOKEN`参考本地部署**

**其他，参考本地部署**

## 本地开发

### 启动前端项目

```ini
# 开发前准备，确保自己的Nodejs版本在18以及以上，全局安装了pnpm

# clone 前端项目
git clone git@github.com:hp0912/wechat-robot-admin-frontend.git

# 进入项目目录
cd wechat-robot-admin-frontend

# 安装依赖
pnpm install

# 生成类型文件
pnpm run build-types

# 启动项目
pnpm run dev
```

### 启动后端项目

```ini
# clone 机器人管理后台后端项目
git clone git@github.com:hp0912/wechat-robot-admin-backend.git

# 进入项目目录
cd wechat-robot-admin-backend

# 下载依赖，翻墙的话会快一点 -> export https_proxy=http://127.0.0.1:7897 http_proxy=http://127.0.0.1:7897 all_proxy=socks5://127.0.0.1:7897
go mod download

# 指定开发模式，这里是mac，win设置环境变量的方式自行探索
export GO_ENV=dev

# 将根目录下的 .env.example 文件复制一份，复制后的文件的文件名改为 .env，按注释说明修改环境变量

# 启动项目
go run main.go
```

### 启动机器人客户端

```ini
# clone 机器人管理后台后端项目
git clone git@github.com:hp0912/wechat-robot-client.git

# 进入项目目录
cd wechat-robot-client

# 下载依赖，翻墙的话会快一点 -> export https_proxy=http://127.0.0.1:7897 http_proxy=http://127.0.0.1:7897 all_proxy=socks5://127.0.0.1:7897
go mod download

# 指定开发模式，这里是mac，win设置环境变量的方式自行探索
export GO_ENV=dev

# 将根目录下的 .env.example 文件复制一份，复制后的文件的文件名改为 .env，按注释说明修改环境变量

# 启动项目
go run main.go
```

### 启动机器人服务端

```yml
services:
  ipad-test:
    image: registry.cn-shenzhen.aliyuncs.com/houhou/wechat-ipad:latest
    container_name: ipad-test
    restart: always
    networks:
      - wechat-robot
    ports:
      - '3010:9000'
    environment:
      WECHAT_PORT: 9000
      REDIS_HOST: wechat-admin-redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: 123456
      REDIS_DB: 0
      WECHAT_CLIENT_HOST: 127.0.0.1:9001
```

```ini
# 机器人服务端，也就是iPad协议，不提供源码，可以通过docker镜像启动，上面是一个 docker-compose.yml 示例

# 向宿主机暴露3010端口，和机器人客户端的 WECHAT_SERVER_HOST 环境变量是相对应的

# WECHAT_CLIENT_HOST REDIS_DB 和机器人客户端环境变量相对应

# redis db 地址、密码别写错了
```

## 官方交流群

<img src="https://img.houhoukang.com/char-room-qrcode.jpg" alt="官方交流群" width="300">