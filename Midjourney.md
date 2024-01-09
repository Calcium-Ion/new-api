# Midjourney Proxy API文档

**简介**:Midjourney Proxy API文档

## 模型价格设置（在设置-运营设置-模型固定价格设置中设置）

```json
{
  "gpt-4-gizmo-*": 0.1,
  "mj_imagine": 0.1,
  "mj_variation": 0.1,
  "mj_reroll": 0.1,
  "mj_blend": 0.1,
  "mj_describe": 0.05,
  "mj_upscale": 0.05
}
```

## 渠道设置

### 对接 midjourney-proxy
1. 部署Midjourney-Proxy，并配置好midjourney账号等（强烈建议设置密钥），[项目地址](https://github.com/novicezk/midjourney-proxy)
2. 在渠道管理中添加渠道，渠道类型选择Midjourney Proxy，模型选择midjourney
3. 地址填写midjourney-proxy部署的地址，例如：http://localhost:8080
4. 密钥填写midjourney-proxy的密钥，如果没有设置密钥，可以随便填

### 对接上游new api
1. 在渠道管理中添加渠道，渠道类型选择Midjourney Proxy，模型选择midjourney
2. 地址填写上游new api的地址，例如：http://localhost:8080
3. 密钥填写上游new api的密钥

## 任务提交

### 绘图变化

**接口地址**:`/mj/submit/change`

**请求方式**:`POST`

**请求数据类型**:`application/json`

**响应数据类型**:`*/*`

**接口描述**:

**请求示例**:

```javascript
{
    "action"
:
    "UPSCALE",
        "index"
:
    1,
        "notifyHook"
:
    "",
        "state"
:
    "",
        "taskId"
:
    "1320098173412546"
}
```

**请求参数**:

| 参数名称                   | 参数说明                                                                  | 请求类型 | 是否必须  | 数据类型           | schema   |
|------------------------|-----------------------------------------------------------------------|------|-------|----------------|----------|
| changeDTO              | changeDTO                                                             | body | true  | 变化任务提交参数       | 变化任务提交参数 |
| &emsp;&emsp;action     | UPSCALE(放大); VARIATION(变换); REROLL(重新生成),可用值:UPSCALE,VARIATION,REROLL |      | true  | string         |          |
| &emsp;&emsp;index      | 序号(1~4), action为UPSCALE,VARIATION时必传                                  |      | false | integer(int32) |          |
| &emsp;&emsp;notifyHook | 回调地址, 为空时使用全局notifyHook                                               |      | false | string         |          |
| &emsp;&emsp;state      | 自定义参数                                                                 |      | false | string         |          |
| &emsp;&emsp;taskId     | 任务ID                                                                  |      | true  | string         |          |

**响应状态**:

| 状态码 | 说明           | schema |
|-----|--------------|--------| 
| 200 | OK           | 提交结果   |
| 201 | Created      |        |
| 401 | Unauthorized |        |
| 403 | Forbidden    |        |
| 404 | Not Found    |        |

**响应参数**:

| 参数名称        | 参数说明                                      | 类型             | schema         |
|-------------|-------------------------------------------|----------------|----------------| 
| code        | 状态码: 1(提交成功), 21(已存在), 22(排队中), other(错误) | integer(int32) | integer(int32) |
| description | 描述                                        | string         |                |
| properties  | 扩展字段                                      | object         |                |
| result      | 任务ID                                      | string         |                |

**响应示例**:

```javascript
{
    "code"
:
    1,
        "description"
:
    "提交成功",
        "properties"
:
    {
    }
,
    "result"
:
    1320098173412546
}
```

### 提交Imagine任务

**接口地址**:`/mj/submit/imagine`

**请求方式**:`POST`

**请求数据类型**:`application/json`

**响应数据类型**:`*/*`

**接口描述**:

**请求示例**:

```javascript
{
    "base64"
:
    "",
        "notifyHook"
:
    "",
        "prompt"
:
    "Cat",
        "state"
:
    ""
}
```

**请求参数**:

| 参数名称                   | 参数说明                    | 请求类型 | 是否必须  | 数据类型        | schema      |
|------------------------|-------------------------|------|-------|-------------|-------------|
| imagineDTO             | imagineDTO              | body | true  | Imagine提交参数 | Imagine提交参数 |
| &emsp;&emsp;base64     | 垫图base64                |      | false | string      |             |
| &emsp;&emsp;notifyHook | 回调地址, 为空时使用全局notifyHook |      | false | string      |             |
| &emsp;&emsp;prompt     | 提示词                     |      | true  | string      |             |
| &emsp;&emsp;state      | 自定义参数                   |      | false | string      |             |

**响应状态**:

| 状态码 | 说明           | schema |
|-----|--------------|--------| 
| 200 | OK           | 提交结果   |
| 201 | Created      |        |
| 401 | Unauthorized |        |
| 403 | Forbidden    |        |
| 404 | Not Found    |        |

**响应参数**:

| 参数名称        | 参数说明                                      | 类型             | schema         |
|-------------|-------------------------------------------|----------------|----------------| 
| code        | 状态码: 1(提交成功), 21(已存在), 22(排队中), other(错误) | integer(int32) | integer(int32) |
| description | 描述                                        | string         |                |
| properties  | 扩展字段                                      | object         |                |
| result      | 任务ID                                      | string         |                |

**响应示例**:

```javascript
{
    "code"
:
    1,
        "description"
:
    "提交成功",
        "properties"
:
    {
    }
,
    "result"
:
    1320098173412546
}
```

## 任务查询

### 指定ID获取任务

**接口地址**:`/mj/task/{id}/fetch`

**请求方式**:`GET`

**请求数据类型**:`application/x-www-form-urlencoded`

**响应数据类型**:`*/*`

**接口描述**:

**请求参数**:

| 参数名称 | 参数说明 | 请求类型 | 是否必须  | 数据类型   | schema |
|------|------|------|-------|--------|--------|
| id   | 任务ID | path | false | string |        |

**响应状态**:

| 状态码 | 说明           | schema |
|-----|--------------|--------| 
| 200 | OK           | 任务     |
| 401 | Unauthorized |        |
| 403 | Forbidden    |        |
| 404 | Not Found    |        |

**响应参数**:

| 参数名称        | 参数说明                                                     | 类型             | schema         |
|-------------|----------------------------------------------------------|----------------|----------------| 
| action      | 可用值:IMAGINE,UPSCALE,VARIATION,REROLL,DESCRIBE,BLEND      | string         |                |
| description | 任务描述                                                     | string         |                |
| failReason  | 失败原因                                                     | string         |                |
| finishTime  | 结束时间                                                     | integer(int64) | integer(int64) |
| id          | 任务ID                                                     | string         |                |
| imageUrl    | 图片url                                                    | string         |                |
| progress    | 任务进度                                                     | string         |                |
| prompt      | 提示词                                                      | string         |                |
| promptEn    | 提示词-英文                                                   | string         |                |
| startTime   | 开始执行时间                                                   | integer(int64) | integer(int64) |
| state       | 自定义参数                                                    | string         |                |
| status      | 任务状态,可用值:NOT_START,SUBMITTED,IN_PROGRESS,FAILURE,SUCCESS | string         |                |
| submitTime  | 提交时间                                                     | integer(int64) | integer(int64) |

**响应示例**:

```javascript
{
    "action"
:
    "",
        "description"
:
    "",
        "failReason"
:
    "",
        "finishTime"
:
    0,
        "id"
:
    "",
        "imageUrl"
:
    "",
        "progress"
:
    "",
        "prompt"
:
    "",
        "promptEn"
:
    "",
        "startTime"
:
    0,
        "state"
:
    "",
        "status"
:
    "",
        "submitTime"
:
    0
}
```