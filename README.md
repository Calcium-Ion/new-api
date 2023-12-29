
# New API

> [!NOTE]
> 本项目为开源项目，在[One API](https://github.com/songquanpeng/one-api)的基础上进行二次开发，感谢原作者的无私奉献。 
> 使用者必须在遵循 OpenAI 的[使用条款](https://openai.com/policies/terms-of-use)以及**法律法规**的情况下使用，不得用于非法用途。


> [!WARNING]
> 本项目为个人学习使用，不保证稳定性，且不提供任何技术支持，使用者必须在遵循 OpenAI 的使用条款以及法律法规的情况下使用，不得用于非法用途。  
> 根据[《生成式人工智能服务管理暂行办法》](http://www.cac.gov.cn/2023-07/13/c_1690898327029107.htm)的要求，请勿对中国地区公众提供一切未经备案的生成式人工智能服务。

> [!NOTE]
> 最新版Docker镜像 calciumion/new-api:latest  
> 更新指令 docker run --rm -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower -cR

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
7. 支持 gpt-4-1106-vision-preview，dall-e-3，tts-1
8. 支持第三方模型 **gps** （gpt-4-gizmo-*），在渠道中添加自定义模型gpt-4-gizmo-*即可
9. 兼容原版One API的数据库，可直接使用原版数据库（one-api.db）
10. 支持模型按次数收费，可在 系统设置-运营设置 中设置
11. 支持gemini-pro，gemini-pro-vision模型
12. 支持渠道**加权随机**

## 部署
### 基于 Docker 进行部署
```shell
# 使用 SQLite 的部署命令：
docker run --name new-api -d --restart always -p 3000:3000 -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest
# 使用 MySQL 的部署命令，在上面的基础上添加 `-e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi"`，请自行修改数据库连接参数。
# 例如：
docker run --name new-api -d --restart always -p 3000:3000 -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest
```

## 交流群
<img src="https://github.com/Calcium-Ion/new-api/assets/61247483/de536a8a-0161-47a7-a0a2-66ef6de81266" width="500">

## 界面截图
![image](https://github.com/Calcium-Ion/new-api/assets/61247483/d1ac216e-0804-4105-9fdc-66b35022d861)

![image](https://github.com/Calcium-Ion/new-api/assets/61247483/3ca0b282-00ff-4c96-bf9d-e29ef615c605)  
![image](https://github.com/Calcium-Ion/new-api/assets/61247483/f4f40ed4-8ccb-43d7-a580-90677827646d)  
![image](https://github.com/Calcium-Ion/new-api/assets/61247483/90d7d763-6a77-4b36-9f76-2bb30f18583d)
![image](https://github.com/Calcium-Ion/new-api/assets/61247483/e414228a-3c35-429a-b298-6451d76d9032)
夜间模式  
![image](https://github.com/Calcium-Ion/new-api/assets/61247483/5b3228e8-2556-44f7-97d6-4f8d8ee6effa)  
![image](https://github.com/Calcium-Ion/new-api/assets/61247483/af9a07ee-5101-4b3d-8bd9-ae21a4fd7e9e)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Calcium-Ion/new-api&type=Date)](https://star-history.com/#Calcium-Ion/new-api&Date)
