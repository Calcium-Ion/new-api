<div align="center">

![new-api](/web/public/logo.png)

# New API

🍥 Next Generation LLM Gateway and AI Asset Management System

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

## 📝 Project Description

> [!NOTE]  
> This is an open-source project developed based on [One API](https://github.com/songquanpeng/one-api)

> [!IMPORTANT]  
> - Users must comply with OpenAI's [Terms of Use](https://openai.com/policies/terms-of-use) and relevant laws and regulations. Not to be used for illegal purposes.
> - This project is for personal learning only. Stability is not guaranteed, and no technical support is provided.

## ✨ Key Features

1. 🎨 New UI interface (some interfaces pending update)
2. 🌍 Multi-language support (work in progress)
3. 🎨 Added [Midjourney-Proxy(Plus)](https://github.com/novicezk/midjourney-proxy) interface support, [Integration Guide](Midjourney.md)
4. 💰 Online recharge support, configurable in system settings:
    - [x] EasyPay
5. 🔍 Query usage quota by key:
    - Works with [neko-api-key-tool](https://github.com/Calcium-Ion/neko-api-key-tool)
6. 📑 Configurable items per page in pagination
7. 🔄 Compatible with original One API database (one-api.db)
8. 💵 Support per-request model pricing, configurable in System Settings - Operation Settings
9. ⚖️ Support channel **weighted random** selection
10. 📈 Data dashboard (console)
11. 🔒 Configurable model access per token
12. 🤖 Telegram authorization login support:
    1. System Settings - Configure Login Registration - Allow Telegram Login
    2. Send /setdomain command to [@Botfather](https://t.me/botfather)
    3. Select your bot, then enter http(s)://your-website/login
    4. Telegram Bot name is the bot username without @
13. 🎵 Added [Suno API](https://github.com/Suno-API/Suno-API) interface support, [Integration Guide](Suno.md)
14. 🔄 Support for Rerank models, compatible with Cohere and Jina, can integrate with Dify, [Integration Guide](Rerank.md)
15. ⚡ **[OpenAI Realtime API](https://platform.openai.com/docs/guides/realtime/integration)** - Support for OpenAI's Realtime API, including Azure channels
16. 🧠 Support for setting reasoning effort through model name suffix:
    - Add suffix `-high` to set high reasoning effort (e.g., `o3-mini-high`)
    - Add suffix `-medium` to set medium reasoning effort
    - Add suffix `-low` to set low reasoning effort

## Model Support
This version additionally supports:
1. Third-party model **gps** (gpt-4-gizmo-*)
2. [Midjourney-Proxy(Plus)](https://github.com/novicezk/midjourney-proxy) interface, [Integration Guide](Midjourney.md)
3. Custom channels with full API URL support
4. [Suno API](https://github.com/Suno-API/Suno-API) interface, [Integration Guide](Suno.md)
5. Rerank models, supporting [Cohere](https://cohere.ai/) and [Jina](https://jina.ai/), [Integration Guide](Rerank.md)
6. Dify

You can add custom models gpt-4-gizmo-* in channels. These are third-party models and cannot be called with official OpenAI keys.

## Additional Configurations Beyond One API
- `GENERATE_DEFAULT_TOKEN`: Generate initial token for new users, default `false`
- `STREAMING_TIMEOUT`: Set streaming response timeout, default 60 seconds
- `DIFY_DEBUG`: Output workflow and node info to client for Dify channel, default `true`
- `FORCE_STREAM_OPTION`: Override client stream_options parameter, default `true`
- `GET_MEDIA_TOKEN`: Calculate image tokens, default `true`
- `GET_MEDIA_TOKEN_NOT_STREAM`: Calculate image tokens in non-stream mode, default `true`
- `UPDATE_TASK`: Update async tasks (Midjourney, Suno), default `true`
- `GEMINI_MODEL_MAP`: Specify Gemini model versions (v1/v1beta), format: "model:version", comma-separated
- `COHERE_SAFETY_SETTING`: Cohere model [safety settings](https://docs.cohere.com/docs/safety-modes#overview), options: `NONE`, `CONTEXTUAL`, `STRICT`, default `NONE`
- `GEMINI_VISION_MAX_IMAGE_NUM`: Gemini model maximum image number, default `16`, set to `-1` to disable
- `MAX_FILE_DOWNLOAD_MB`: Maximum file download size in MB, default `20`
- `CRYPTO_SECRET`: Encryption key for encrypting database content
- `AZURE_DEFAULT_API_VERSION`: Azure channel default API version, if not specified in channel settings, use this version, default `2024-12-01-preview`

## Deployment
> [!TIP]
> Latest Docker image: `calciumion/new-api:latest`  
> Default account: root, password: 123456  
> Update command:
> ```
> docker run --rm -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower -cR
> ```

### Multi-Server Deployment
- Must set `SESSION_SECRET` environment variable, otherwise login state will not be consistent across multiple servers.
- If using a public Redis, must set `CRYPTO_SECRET` environment variable, otherwise Redis content will not be able to be obtained in multi-server deployment.

### Requirements
- Local database (default): SQLite (Docker deployment must mount `/data` directory)
- Remote database: MySQL >= 5.7.8, PgSQL >= 9.6

### Docker Deployment
### Using Docker Compose (Recommended)
```shell
# Clone project
git clone https://github.com/Calcium-Ion/new-api.git
cd new-api
# Edit docker-compose.yml as needed
# Start
docker-compose up -d
```

### Direct Docker Image Usage
```shell
# SQLite deployment:
docker run --name new-api -d --restart always -p 3000:3000 -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest
# MySQL deployment (add -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi"), modify database connection parameters as needed
# Example:
docker run --name new-api -d --restart always -p 3000:3000 -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest
```

## Channel Retry
Channel retry is implemented, configurable in `Settings->Operation Settings->General Settings`. **Cache recommended**.  
First retry uses same priority, second retry uses next priority, and so on.

### Cache Configuration
1. `REDIS_CONN_STRING`: Use Redis as cache
    + Example: `REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `MEMORY_CACHE_ENABLED`: Enable memory cache, default `false`
    + Example: `MEMORY_CACHE_ENABLED=true`

### Why Some Errors Don't Retry
Error codes 400, 504, 524 won't retry
### To Enable Retry for 400
In `Channel->Edit`, set `Status Code Override` to:
```json
{
  "400": "500"
}
```

## Integration Guides
- [Midjourney Integration](Midjourney.md)
- [Suno Integration](Suno.md)

## Related Projects
- [One API](https://github.com/songquanpeng/one-api): Original project
- [Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy): Midjourney interface support
- [chatnio](https://github.com/Deeptrain-Community/chatnio): Next-gen AI B/C solution
- [neko-api-key-tool](https://github.com/Calcium-Ion/neko-api-key-tool): Query usage quota by key

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Calcium-Ion/new-api&type=Date)](https://star-history.com/#Calcium-Ion/new-api&Date)