
# Neko API

> **Note**
> 本项目为开源项目，在[One API](https://github.com/songquanpeng/one-api)的基础上进行二次开发，感谢原作者的无私奉献。 
> 使用者必须在遵循 OpenAI 的[使用条款](https://openai.com/policies/terms-of-use)以及**法律法规**的情况下使用，不得用于非法用途。


> **Warning**
> 本项目为个人学习使用，不保证稳定性，且不提供任何技术支持，使用者必须在遵循 OpenAI 的使用条款以及法律法规的情况下使用，不得用于非法用途。  
> 根据[《生成式人工智能服务管理暂行办法》](http://www.cac.gov.cn/2023-07/13/c_1690898327029107.htm)的要求，请勿对中国地区公众提供一切未经备案的生成式人工智能服务。

> **Note**
> 最新版Docker镜像 calciumion/neko-api:main

## 此分叉版本的主要变更
1. 全新的UI界面（部分界面还待更新）
2. 添加[Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy)接口的支持：
    + [x] /mj/submit/imagine
    + [x] /mj/submit/change
    + [x] /mj/submit/blend
    + [x] /mj/submit/describe
    + [x] /mj/image/{id} （通过此接口获取图片，**请必须在系统设置中填写服务器地址！！**）
    + [x] /mj/task/{id}/fetch （此接口返回的图片地址为经过One API转发的地址）
3. 支持在线充值功能，可在系统设置中设置，当前支持的支付接口：
    + [x] 易支付
4. 支持用key查询使用额度:
    + 配合项目[neko-api-key-tool](https://github.com/Calcium-Ion/neko-api-key-tool)可实现用key查询使用情况，方便二次分销
5. 渠道显示已使用额度，支持指定组织访问
6. 分页支持选择每页显示数量

