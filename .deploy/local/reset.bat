@echo off

REM 停止并移除容器
docker compose down
docker compose -f docker-compose2.yml down

REM 删除数据文件夹和文件
rmdir /s /q wechat_admin_mysql_data
rmdir /s /q wechat_admin_redis_data
rmdir /s /q wechat-server

REM 拉取最新镜像
docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat-robot-admin-frontend:latest
docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat-robot-admin-backend:latest
docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat-robot-client:latest
docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat-ipad:latest

REM 重启服务
docker compose -f docker-compose2.yml up -d

pause