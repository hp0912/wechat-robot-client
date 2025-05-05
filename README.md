# 微信机器人客户端

```vim
docker run -it --rm -v /Users/zuihoudeqingyu/Git/wechat/wechat-robot-client:/go/release \
-e ROBOT_ID=1  \
-e ROBOT_CODE=localhost  \
-e ROBOT_START_TIMEOUT=60  \
-e MYSQL_DRIVER=mysql  \
-e MYSQL_HOST=host.docker.internal  \
-e MYSQL_PORT=3306  \
-e MYSQL_USER=root  \
-e MYSQL_PASSWORD=houhou  \
-e MYSQL_ADMIN_DB=robot_admin  \
-e MYSQL_DB=root  \
-e MYSQL_SCHEMA=public  \
-e GOPROXY=https://goproxy.cn,direct \
-w /go/release \
-p 9002:9002  \
--entrypoint sh \
registry.cn-shenzhen.aliyuncs.com/houhou/silk-base:latest
```
