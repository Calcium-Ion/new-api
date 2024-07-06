# Suno API文档

**简介**:Suno API文档

## 接口列表
支持的接口如下：
+ [x] /suno/submit/music
+ [x] /suno/submit/lyrics
+ [x] /suno/fetch
+ [x] /suno/fetch/:id

## 模型列表

### Suno API支持

- suno_music (自定义模式、灵感模式、续写)
- suno_lyrics (生成歌词)


## 模型价格设置（在设置-运营设置-模型固定价格设置中设置）
```json
{
  "suno_music": 0.3,
  "suno_lyrics": 0.01
}
```

## 渠道设置

### 对接 Suno API

1.
部署 Suno API，并配置好suno账号等（强烈建议设置密钥），[项目地址](https://github.com/Suno-API/Suno-API)

2. 在渠道管理中添加渠道，渠道类型选择**Suno API**
   ，模型请参考上方模型列表
3. **代理**填写 Suno API 部署的地址，例如：http://localhost:8080
4. 密钥填写 Suno API 的密钥，如果没有设置密钥，可以随便填

### 对接上游new api

1. 在渠道管理中添加渠道，渠道类型选择**Suno API**，或任意类型，只需模型包含上方模型列表的模型
2. **代理**填写上游new api的地址，例如：http://localhost:3000
3. 密钥填写上游new api的密钥