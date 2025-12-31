# 更新日志

## [2.2.0] - 2025/12/31

### 破坏性更新

- 表结构新增和更新，请按照 [SQL 升级脚本](https://github.com/hp0912/wechat-robot-admin-backend/blob/main/template/2_2_0.sql)进行升级

### 新特性

- 新增表情包提取 MCP 工具

- 新增群成员管理功能，将群成员设置为管理员 / 将群成员加入黑名单(黑名单的成员不会触发 AI 交互)

### 体验性优化

- 优化聊天记录查询性能

- 优化 MCP 协议，新增 MCP 私有协议，用于发送聊天记录、图片、视频、语音等等

- 优化群聊总结提示词，新增提取群聊分享的链接资源

- Mac 自动过滑块改为手动过滑块，提升准确性

### BUG 修复

- 修复引用消息丢失了部分上下文的问题

- 修复群成员最后活跃时间不正确的问题

## [2.1.0] - 2025/11/23

### 体验性优化

- 图片 / 视频改成分片发送，解决发送大视频内存爆炸的问题

## [2.0.0] - 2025/11/15

### 破坏性更新

- `wechat-robot-admin-backend`服务新增了一个必填环境变量`UUID_URL`，可填写为`http://wechat-slider:9000`

- 表结构新增和更新，请按照 [SQL 升级脚本](https://github.com/hp0912/wechat-robot-admin-backend/blob/main/template/2_0_0.sql)进行升级

- 需要更新`wechat-slider`服务的镜像，执行 `docker pull registry.cn-shenzhen.aliyuncs.com/houhou/wechat/wechat-slider-base:latest`

- 移除 `jimeng-free-api` 服务，执行 `docker compose rm -s -f jimeng-free-api` 或者 `docker-compose rm -s -f jimeng-free-api`，哪个能用用哪个

- 拉取最新代码获取最新`docker-compose.yml`文件，根据文件内容，手动拉一遍镜像，成功率更高。

### 新特性

- 完整的 MCP 协议支持

- 开放微信消息 Webhook 回调

### 体验性优化

- iPad 协议新增支持 pprof 监控(启用方法: 机器人详情界面，更新镜像下拉框 -> 删除服务端容器 -> 创建服务端容器 (启用pprof))

- 即梦逆向 api 由 [jimeng-free-api](https://github.com/LLM-Red-Team/jimeng-free-api) 迁移到 [jimeng-api](https://github.com/iptag/jimeng-api)，以支持更多功能

- 优化机器人管理后台前端项目本地Docker构建(本地构建使用 dev.Dockerfile)

### BUG 修复

- 修复 iPad 协议提取 ticket 异常的问题

## [1.6.0] - 2025/10/12

### 体验性优化

- 提高抖音视频解析稳定性。(破坏性更新： 抖音解析的 API 由免费 API 改为收费 API，原先的环境变量`THIRD_PARTY_API_KEY`记得保持余额充足，一分钱解析10次。附：[充值链接](https://api.pearktrue.cn/dashboard/profile))

## [1.5.2] - 2025/10/01

### 新特性

- 消息图片自动上传 OSS (需要执行数据库升级脚本[https://github.com/hp0912/wechat-robot-admin-backend/blob/1.3.2/template/1_3_2.sql](https://github.com/hp0912/wechat-robot-admin-backend/blob/1.3.2/template/1_3_2.sql))

## [1.5.1] - 2025/09/27

### 体验性优化

- 支持登录设备迁移 (导出登录信息、导入登录信息)

## [1.5.0] - 2025/09/27

### 体验性优化

- 显示当前登录设备类型和微信版本。

## [1.4.3] - 2025/09/22

### 新特性

- iPad 伪装登录。

## [1.4.2] - 2025/09/21

### BUG 修复

- 修复点歌接口挂了的问题

- 修复发送AI消息获取音色接口挂了的问题

- 修复扫码登录UUID检测异常的问题

## [1.4.1] - 2025/09/21

### 新特性

- Mac 扫码登录支持自动过滑块。

> 本次更新包含破坏性更新
>
> `wechat-robot-admin-backend`服务需要新增两个环境变量，否则服务会启动失败
>
> - SLIDER_SERVER_BASE_URL=http://wechat-slider:9000
>
> - SLIDER_TOKEN=xxxxxxx # 滑块验证码服务密钥，请加入官方交流群获取
>

## [1.4.0] - 2025/09/20

### 新特性

- ~~Mac 扫码登录支持手动过滑块。~~

> 本次更新包含破坏性更新
>
> `wechat-robot-admin-backend`服务需要新增两个环境变量，否则服务会启动失败
>
> - ~~SLIDER_VERIFY_URL=http://wechat-slider:9000/api/v1/slider-verify-html~~
>
> - ~~SLIDER_VERIFY_SUBMIT_URL=http://wechat-slider:9000/api/v1/security-verify~~
>

## [1.3.0] - 2025/09/19

### 体验性优化

- 抖音视频会同时发送链接和视频。

### BUG 修复

- 修复 Mac 登录异常的问题。

## [1.2.0] - 2025/09/10

### 体验性优化

- 抖音视频由手动出发改为自动触发。

## [1.1.15] - 2025/09/09

### 体验性优化

- 抖音视频解析由直接发送视频改为发送卡片链接。

## [1.1.14] - 2025/09/06

### 体验性优化

- 公众号如果没有手动开启AI聊天，默认不开启 (原来会继承全局 AI 设置)。

- 群聊总结改为以聊天记录的形式发送，避免内容过多被折叠。

## [1.1.13] - 2025/09/04

### BUG 修复

- 修复因为每日早报上游接口数据结构变化导致获取每日早报失败的问题

- 修复AI机器人在私聊场景会响应自己发送(从其他设备发送)的消息的问题。

## [1.1.12] - 2025/08/29

### 新特性

- 支持通过登录密钥登录机器人管理后台

## [1.1.11] - 2025/08/19

### 新特性

- 支持多种登录方式(iPad、Windows微信、车载微信、Mac微信)

- 支持Data62登录

- 支持A16登录

### 修改

- 优化创建机器人docker容器流程，如果docker镜像还没拉取过，会自动拉取 (wechat-robot-admin-backend)

## [1.1.10] - 2025/08/18

### 新特性

- 新增文本消息群发接口

## [1.1.9] - 2025/08/16

### 新特性

- 支持发送文件消息，流式发送，避免发送超大文件时，内存溢出