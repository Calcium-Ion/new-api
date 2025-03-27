# One API 接口文档 (OpenAPI 3.0)

## 概述

One API 是一个统一的API网关，支持多个LLM模型和服务集成。本文档基于OpenAPI 3.0规范提供了API的详细说明和使用方法。

## 访问OpenAPI文档

项目集成了Swagger UI，方便直观地浏览和测试API。访问方式：

1. 启动One API服务
2. 打开浏览器，访问 `http://localhost:3000/swagger/index.html`

## API认证

除了少数公开接口外，大多数API都需要认证才能访问：

1. **Bearer认证**: 通过`Authorization: Bearer <your_token>`方式进行认证
2. **会话认证**: 通过登录后的Cookie/会话认证进行访问

这两种认证方式已在OpenAPI文档中定义。

## 主要接口分类

### 用户管理

- `/api/user/login` - 用户登录
- `/api/user/register` - 用户注册 
- `/api/user/logout` - 退出登录
- `/api/user/self` - 获取当前用户信息
- `/api/user/token` - 生成访问令牌

### 渠道管理

- `/api/channel` - 获取/添加/更新渠道
- `/api/channel/{id}` - 获取/删除特定渠道
- `/api/channel/search` - 搜索渠道
- `/api/channel/test/{id}` - 测试渠道可用性

### 模型管理

- `/api/models` - 获取可用模型列表

### 令牌管理

- `/api/token` - 管理访问令牌

## 响应格式

所有API响应都遵循统一的JSON格式：

```json
{
  "success": true,  // 操作是否成功
  "message": "",    // 提示信息，成功时通常为空
  "data": {}        // 返回的数据，根据接口不同而变化
}
```

错误响应：

```json
{
  "success": false,         // 操作失败
  "message": "错误信息",    // 具体错误描述
}
```

## 分页查询

支持分页的接口通常接受以下参数：

- `p` - 页码，从0开始
- `page_size` - 每页条目数

## OpenAPI 3.0的改进

相比Swagger 2.0，OpenAPI 3.0规范有以下改进：

1. 更好的组件复用 - 使用`components`替代`definitions`
2. 增强的安全定义 - 更灵活的安全配置
3. 更丰富的内容类型支持 - 使用`content`定义请求和响应体
4. 更好的参数描述能力 - 使用`schema`更清晰地描述参数

## 更多帮助

详细的API信息请参考Swagger UI，如有问题可以提交GitHub Issue。 