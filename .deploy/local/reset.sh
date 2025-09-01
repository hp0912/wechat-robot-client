docker compose down
docker compose -f docker-compose2.yml down

rm -rf wechat_admin_mysql_data wechat_admin_redis_data wechat-server

docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat-robot-admin-frontend:latest

docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat-robot-admin-backend:latest

docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat-robot-client:latest

docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat-ipad:latest

docker compose -f docker-compose2.yml up -d