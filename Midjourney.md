# Midjourney Proxy API文档

**简介**:Midjourney Proxy API文档

## 模型价格设置（在设置-运营设置-模型固定价格设置中设置）

```json
{
  "gpt-4-gizmo-*": 0.1,
  "mj_imagine":     0.1,
  "mj_variation":   0.1,
  "mj_reroll":      0.1,
  "mj_blend":       0.1,
  "mj_inpaint":     0.1,
  "mj_zoom":        0.1,
  "mj_inpaint_pre": 0,
  "mj_describe":    0.05,
  "mj_upscale":     0.05,
  "swap_face":     0.05
}
```

## 渠道设置

### 对接 midjourney-proxy(plus)
1. 部署Midjourney-Proxy，并配置好midjourney账号等（强烈建议设置密钥），[项目地址](https://github.com/novicezk/midjourney-proxy)
2. 在渠道管理中添加渠道，渠道类型选择**Midjourney Proxy**，如果是plus版本选择**Midjourney Proxy Plus**，模型选择midjourney，如果有换脸模型，可以选择swap_face
3. 地址填写midjourney-proxy部署的地址，例如：http://localhost:8080
4. 密钥填写midjourney-proxy的密钥，如果没有设置密钥，可以随便填

### 对接上游new api
1. 在渠道管理中添加渠道，渠道类型选择**Midjourney Proxy Plus**，模型选择midjourney，如果有换脸模型，可以选择swap_face
2. 地址填写上游new api的地址，例如：http://localhost:3000
3. 密钥填写上游new api的密钥