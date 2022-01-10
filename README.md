# biligo-live-ws

以 [bili-go](https://github.com/iyear/biligo-live) 为核心, 基于 JSON string 序列化 的 B站 WebSocket 监控服务器。

## 简介

根据个人开发经验，每次开发与B站直播WS相关的项目时，都要烦恼自己所使用的编程语言有没有相关的B站WS处理库，如果没有的话更要自己实作一个，当中的 binary 处理和解压程序可谓相当麻烦。

### 连线即用

打开后，透过 POST 请求输入你的订阅房间列表，然后连入WS后即可开始接收JSON数据，**无需binary处理**。

你也可以连入 WebSocket 之后才输入你的订阅

### 即时增减

可以透过 PUT 请求即时新增和删除批量房间号，此举**不需要**透过重连 WebSocket 来刷新

### 无需后端

透过直接存取公共 API 地址，可直接在前端获取B站直播数据而无需自架或开发后端。

## 直接使用

目前的公共 API 地址: 

- https://blive.ericlamm.xyz/ (位置: 香港)
- https://blive-jp.ericlamm.xyz/ (位置: 大阪)

执行 GET / 后 显示

```json
{
  "status": "working"
}

```
则代表服务正常运行。

可供测试的前端地址: https://eric2788.github.io/biligo-live-ws

### 使用方式

#### 头规格

| Header        | Value                                     |
|---------------|-------------------------------------------|
| Content-Type  | application/x-www-form-urlencoded         |
| Authorization | 非必填，辨识ID，一个IP多程序用的时候防止混淆；不填则用 `anonymous` |

#### 注意

如果标明了 `Authorization`，则连入 websocket 时需要传入 query string `?id={辨识ID}`

假设你设置订阅时传入头

```json
{
  "Content-Type": "application/x-www-form-urlencoded",
  "Authorization": "abc"
}
```

连入 WS 则需要使用

``
wss://blive.chu77.xyz/ws?id=abc
``

#### 步骤

1. 透过 POST /subscribe 透过 `subscribes` key 递交你的订阅列表 (数组)

    例如
    
    ```bash
    subscribes=545&subscribes=114514
    ```
    
    成功后将返回**成功订阅**(无效的房间会被忽略)的**真实房间**号列表 (数组)
    
    ```json
    [
      573893
    ]
    ```

   **注意，如果订阅后五分钟内没有连入 WebSocket, 将会自动清除订阅列表数据(断线后也会开始计时)**


2. 透过 POST /validate 检查是否一切准备就绪(非必要)

    此请求会检查你的订阅列表是否为空，如果为空将返回 400 错误，否则返回 200


3. 开始透过 后缀为 /ws 的请求连入 WebSocket

   如果成功订阅，连入 WebSocket 的几秒后将会开始接收已经 JSON 序列化的 B站直播 数据

### API 参考

| Path 路径           | Method 方法 | Payload 传入   | Response(200) 返回 | Error 错误                   |
|-------------------|-----------|--------------|------------------|----------------------------|
| /                 | GET       | 无            | 程序是否运行           | 无                          |
| /subscribe        | GET       | 无            | 目前的订阅列表(数组)      | 无                          |
| /subscribe        | POST      | 订阅列表(数组)     | 成功的订阅列表(数组)      | 400 如果輸入列表为空或缺少数值          |
| /subscribe        | DELETE    | 删除订阅列表       | 无                | 无                          |
| /validate         | POST      | 无            | 无                | 400 如果准备未就绪                |
| /subscribe/add    | PUT       | 要新增的批量订阅(数组) | 目前的订阅列表(数组)      | 400 如果輸入列表为空或缺少数值          |
| /subscribe/remove | PUT       | 要删除的批量订阅(数组) | 目前的订阅列表(数组)      | 400 如果輸入列表为空或缺少数值/之前尚未递交订阅 |

### B站直播数据解析

格式如下

| key       | 数值                 | 类型     |
|-----------|--------------------|--------|
| command   | 直播数据指令             | string |
| live_info | 直播房间资讯             | 详见下方   |
| content   | 直播数据原始内容(已转换为json) | object |

直播房间资讯

| key     | 数值     | 类型     |
|---------|--------|--------|
| room_id | 直播房间号  | int64  |
| uid     | 直播用户ID | int64  |
| title   | 直播标题   | int64  |
| name    | 直播名称   | string |
| cover   | 直播封面网址 | string |

**每次开播时都会自动刷新一次直播房间资讯**

### 备注

- 指令为 `HEARTBEAT_REPLY` 的**直播数据原始内容**已被序列化为格式

   ```json
   {
      "popularity": 999999
   }
   ```
   (999999为人气值)


- 直播数据原始内容(content) 如果转换 `object` 失败，将自动转为 `string`

## 私人部署

### Docker
[docker.io](https://hub.docker.com/r/eric1008818/biligo-live-ws) 或 详见 Dockerfile

### Linux / Windows

详见 [Releases](https://github.com/eric2788/biligo-live-ws/releases)

运行参数(非必要)

```bash
./biligo-live-ws 端口
```

端口: 不填则 8080

## 鸣谢

[bili-go](https://github.com/iyear/biligo-live) 作者
