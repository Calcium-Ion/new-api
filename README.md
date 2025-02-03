<p align="right">
   <strong>中文</strong> | <a href="./README.en.md">English</a>
</p>
<div align="center">

![new-api](/web/public/logo.png)

# New API


🍥新一代大模型网关与AI资产管理系统

<a href="https://trendshift.io/repositories/8227" target="_blank"><img src="https://trendshift.io/api/badge/repositories/8227" alt="Calcium-Ion%2Fnew-api | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

<p align="center">
  <a href="https://raw.githubusercontent.com/Calcium-Ion/new-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/Calcium-Ion/new-api?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/Calcium-Ion/new-api/releases/latest">
    <img src="https://img.shields.io/github/v/release/Calcium-Ion/new-api?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/users/Calcium-Ion/packages/container/package/new-api">
    <img src="https://img.shields.io/badge/docker-ghcr.io-blue" alt="docker">
  </a>
  <a href="https://hub.docker.com/r/CalciumIon/new-api">
    <img src="https://img.shields.io/badge/docker-dockerHub-blue" alt="docker">
  </a>
  <a href="https://goreportcard.com/report/github.com/Calcium-Ion/new-api">
    <img src="https://goreportcard.com/badge/github.com/Calcium-Ion/new-api" alt="GoReportCard">
  </a>
</p>
</div>

## 📝 项目说明

> [!NOTE]  
> 本项目为开源项目，在[One API](https://github.com/songquanpeng/one-api)的基础上进行二次开发

> [!IMPORTANT]  
> - 使用者必须在遵循 OpenAI 的[使用条款](https://openai.com/policies/terms-of-use)以及**法律法规**的情况下使用，不得用于非法用途。
> - 本项目仅供个人学习使用，不保证稳定性，且不提供任何技术支持。
> - 根据[《生成式人工智能服务管理暂行办法》](http://www.cac.gov.cn/2023-07/13/c_1690898327029107.htm)的要求，请勿对中国地区公众提供一切未经备案的生成式人工智能服务。

## ✨ 主要特性

1. 🎨 全新的UI界面（部分界面还待更新）
2. 🌍 多语言支持（待完善）
3. 🎨 添加[Midjourney-Proxy(Plus)](https://github.com/novicezk/midjourney-proxy)接口支持，[对接文档](Midjourney.md)
4. 💰 支持在线充值功能，可在系统设置中设置：
    - [x] 易支付
5. 🔍 支持用key查询使用额度：
    - 配合项目[neko-api-key-tool](https://github.com/Calcium-Ion/neko-api-key-tool)可实现用key查询使用
6. 📑 分页支持选择每页显示数量
7. 🔄 兼容原版One API的数据库，可直接使用原版数据库（one-api.db）
8. 💵 支持模型按次数收费，可在 系统设置-运营设置 中设置
9. ⚖️ 支持渠道**加权随机**
10. 📈 数据看板（控制台）
11. 🔒 可设置令牌能调用的模型
12. 🤖 支持Telegram授权登录：
    1. 系统设置-配置登录注册-允许通过Telegram登录
    2. 对[@Botfather](https://t.me/botfather)输入指令/setdomain
    3. 选择你的bot，然后输入http(s)://你的网站地址/login
    4. Telegram Bot 名称是bot username 去掉@后的字符串
13. 🎵 添加 [Suno API](https://github.com/Suno-API/Suno-API)接口支持，[对接文档](Suno.md)
14. 🔄 支持Rerank模型，目前兼容Cohere和Jina，可接入Dify，[对接文档](Rerank.md)
15. ⚡ **[OpenAI Realtime API](https://platform.openai.com/docs/guides/realtime/integration)** - 支持OpenAI的Realtime API，支持Azure渠道
16. 支持使用路由/chat2link 进入聊天界面
17. 🧠 支持通过模型名称后缀设置 reasoning effort：
    - 添加后缀 `-high` 设置为 high reasoning effort (例如: `o3-mini-high`)
    - 添加后缀 `-medium` 设置为 medium reasoning effort (例如: `o3-mini-medium`)
    - 添加后缀 `-low` 设置为 low reasoning effort (例如: `o3-mini-low`)

## 模型支持
此版本额外支持以下模型：
1. 第三方模型 **gps** （gpt-4-gizmo-*）
2. [Midjourney-Proxy(Plus)](https://github.com/novicezk/midjourney-proxy)接口，[对接文档](Midjourney.md)
3. 自定义渠道，支持填入完整调用地址
4. [Suno API](https://github.com/Suno-API/Suno-API) 接口，[对接文档](Suno.md)
5. Rerank模型，目前支持[Cohere](https://cohere.ai/)和[Jina](https://jina.ai/)，[对接文档](Rerank.md)
6. Dify

您可以在渠道中添加自定义模型gpt-4-gizmo-*，此模型并非OpenAI官方模型，而是第三方模型，使用官方key无法调用。

## 比原版One API多出的配置
- `GENERATE_DEFAULT_TOKEN`：是否为新注册用户生成初始令牌，默认为 `false`。
- `STREAMING_TIMEOUT`：设置流式一次回复的超时时间，默认为 60 秒。
- `DIFY_DEBUG`：设置 Dify 渠道是否输出工作流和节点信息到客户端，默认为 `true`。
- `FORCE_STREAM_OPTION`：是否覆盖客户端stream_options参数，请求上游返回流模式usage，默认为 `true`，建议开启，不影响客户端传入stream_options参数返回结果。
- `GET_MEDIA_TOKEN`：是否统计图片token，默认为 `true`，关闭后将不再在本地计算图片token，可能会导致和上游计费不同，此项覆盖 `GET_MEDIA_TOKEN_NOT_STREAM` 选项作用。
- `GET_MEDIA_TOKEN_NOT_STREAM`：是否在非流（`stream=false`）情况下统计图片token，默认为 `true`。
- `UPDATE_TASK`：是否更新异步任务（Midjourney、Suno），默认为 `true`，关闭后将不会更新任务进度。
- `GEMINI_MODEL_MAP`：Gemini模型指定版本(v1/v1beta)，使用"模型:版本"指定，","分隔，例如：-e GEMINI_MODEL_MAP="gemini-1.5-pro-latest:v1beta,gemini-1.5-pro-001:v1beta"，为空则使用默认配置(v1beta)
- `COHERE_SAFETY_SETTING`：Cohere模型[安全设置](https://docs.cohere.com/docs/safety-modes#overview)，可选值为 `NONE`, `CONTEXTUAL`, `STRICT`，默认为 `NONE`。
- `GEMINI_VISION_MAX_IMAGE_NUM`：Gemini模型最大图片数量，默认为 `16`，设置为 `-1` 则不限制。
- `MAX_FILE_DOWNLOAD_MB`: 最大文件下载大小，单位 MB，默认为 `20`。
- `CRYPTO_SECRET`：加密密钥，用于加密数据库内容。
- `AZURE_DEFAULT_API_VERSION`：Azure渠道默认API版本，如果渠道设置中未指定API版本，则使用此版本，默认为 `2024-12-01-preview`
## 部署
> [!TIP]
> 最新版Docker镜像：`calciumion/new-api:latest`  
> 默认账号root 密码123456  
> 更新指令：
> ```
> docker run --rm -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower -cR
> ```

### 多机部署
- 必须设置环境变量 `SESSION_SECRET`，否则会导致多机部署时登录状态不一致。
- 如果公用Redis，必须设置 `CRYPTO_SECRET`，否则会导致多机部署时Redis内容无法获取。

### 部署要求
- 本地数据库（默认）：SQLite（Docker 部署默认使用 SQLite，必须挂载 `/data` 目录到宿主机）
- 远程数据库：MySQL 版本 >= 5.7.8，PgSQL 版本 >= 9.6

### 使用宝塔面板Docker功能部署
安装宝塔面板 (**9.2.0版本**及以上)，前往 [宝塔面板](https://www.bt.cn/new/download.html) 官网，选择正式版的脚本下载安装  
安装后登录宝塔面板，在菜单栏中点击 Docker ，首次进入会提示安装 Docker 服务，点击立即安装，按提示完成安装  
安装完成后在应用商店中找到 **New-API** ，点击安装，配置基本选项 即可完成安装  
[图文教程](BT.md)

### 基于 Docker 进行部署
### 使用 Docker Compose 部署（推荐）
```shell
# 下载项目
git clone https://github.com/Calcium-Ion/new-api.git
cd new-api
# 按需编辑 docker-compose.yml
# 启动
docker-compose up -d
```

### 直接使用 Docker 镜像
```shell
# 使用 SQLite 的部署命令：
docker run --name new-api -d --restart always -p 3000:3000 -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest
# 使用 MySQL 的部署命令，在上面的基础上添加 `-e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi"`，请自行修改数据库连接参数。
# 例如：
docker run --name new-api -d --restart always -p 3000:3000 -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest
```

## 渠道重试
渠道重试功能已经实现，可以在`设置->运营设置->通用设置`设置重试次数，**建议开启缓存**功能。  
如果开启了重试功能，第一次重试使用同优先级，第二次重试使用下一个优先级，以此类推。
### 缓存设置方法
1. `REDIS_CONN_STRING`：设置之后将使用 Redis 作为缓存使用。
    + 例子：`REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `MEMORY_CACHE_ENABLED`：启用内存缓存（如果设置了`REDIS_CONN_STRING`，则无需手动设置），会导致用户额度的更新存在一定的延迟，可选值为 `true` 和 `false`，未设置则默认为 `false`。
    + 例子：`MEMORY_CACHE_ENABLED=true`
### 为什么有的时候没有重试
这些错误码不会重试：400，504，524
### 我想让400也重试
在`渠道->编辑`中，将`状态码复写`改为
```json
{
  "400": "500"
}
```
可以实现400错误转为500错误，从而重试

## Midjourney接口设置文档
[对接文档](Midjourney.md)

## Suno接口设置文档
[对接文档](Suno.md)

## 界面截图
![image](https://github.com/user-attachments/assets/a0dcd349-5df8-4dc8-9acf-ca272b239919)


![image](https://github.com/user-attachments/assets/c7d0f7e1-729c-43e2-ac7c-2cb73b0afc8e)

![image](https://github.com/user-attachments/assets/29f81de5-33fc-4fc5-a5ff-f9b54b653c7c)

![image](https://github.com/user-attachments/assets/4fa53e18-d2c5-477a-9b26-b86e44c71e35)

## 交流群
<img src="https://github.com/user-attachments/assets/9ca0bc82-e057-4230-a28d-9f198fa022e3" width="200">

## 相关项目
- [One API](https://github.com/songquanpeng/one-api)：原版项目
- [Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy)：Midjourney接口支持
- [chatnio](https://github.com/Deeptrain-Community/chatnio)：下一代 AI 一站式 B/C 端解决方案
- [neko-api-key-tool](https://github.com/Calcium-Ion/neko-api-key-tool)：用key查询使用额度

其他基于New API的项目：
- [new-api-horizon](https://github.com/Calcium-Ion/new-api-horizon)：New API高性能优化版，并支持Claude格式
- [VoAPI](https://github.com/VoAPI/VoAPI)：基于New API的闭源项目

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Calcium-Ion/new-api&type=Date)](https://star-history.com/#Calcium-Ion/new-api&Date)
