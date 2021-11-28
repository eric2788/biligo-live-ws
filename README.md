# biligo-live-ws

以 [bili-go](https://github.com/iyear/biligo-live) 为核心, 基于 JSON string 序列化 的 B站 WebSocket 监控服务器。

## 简介

根据个人开发经验，每次开发与B站直播WS相关的项目时，都要烦恼自己所使用的编程语言有没有相关的B站WS处理库，如果没有的话更要自己实作一个，当中的 binary 转换和解压程序可谓相当麻烦。

## 直接使用
**(暂定测试用，之后域名及位置会有所更改)**

目前可供测试的 API 地址: https://blive.chu77.xyz/

执行 GET / 后 显示

```json
{
  "status": "working"
}

```
则代表服务正常运行。

可供测试的前端地址: https://eric2788.github.io/biligo-live-ws

### 开箱即用

打开后，透过 POST 请求输入你的订阅房间列表，然后连入WS后即可开始接收JSON数据。

### 使用方式

#### 规格

| Header      | Value |
| ----------- | ----------- |
| Content-Type      | application/x-www-form-urlencoded       |


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

   如果上述都通过，再连入 WebSocket 的几秒后将会开始接收已经 JSON 序列化的 B站直播 数据

### API 参考

| Path 路径 | Method 方法 | Payload 传入 | Response(200) 返回 | Error 错误 |
| -------- | ----------- | ------------------ | ------------ | ----- |
| /   | GET | 无 | 程序是否运行 | 无 |
| /subscribe | GET | 无 | 目前的订阅列表(数组) | 无 |
| /subscribe | POST | 订阅列表(数组) | 成功的订阅列表(数组) | 400 如果輸入列表为空或缺少数值 |
| /subscribe | DELETE | 删除订阅列表 | 无 | 无 |
| /validate | POST | 无 | 无 | 400 如果准备未就绪 |
| /subscribe/add | PUT | 要新增的批量订阅(数组) | 目前的订阅列表(数组) | 400 如果輸入列表为空或缺少数值 |
| /subscribe/remove | PUT | 要删除的批量订阅(数组) | 目前的订阅列表(数组) | 400 如果輸入列表为空或缺少数值/之前尚未递交订阅 |

### B站直播数据解析

格式如下

| key | 数值 | 类型 |
| ---- | --- | ---- |
| command | 直播数据指令 | string |
| live_info | 直播房间资讯 | 详见下放 |
| content | 直播数据原始内容(已转换为 json) | string |

直播房间资讯

| key | 数值 | 类型 |
| ---- | --- | ---- |
| room_id | 直播房间号 | int64 |
| uid | 直播用户ID | int64 |
| title | 直播标题 | int64 |
| name | 直播名称 | string |
| cover | 直播封面网址 | string |

**每次开播时都会自动刷新一次直播房间资讯**

### 备注

- 指令为 `HEARTBEAT_REPLY` 的**直播数据原始内容**已被序列化为格式

   ```json
   {
      "popularity": 999999
   }
   ```
   (999999为人气值)


- 直播数据原始内容(content) 在 json 反序列化后的数值类型为 string, 你需要再一次反序列化以转换为 object


## 私人部署

### Docker
详见 Dockerfile

### Linux / Windows

详见 [Releases](https://github.com/eric2788/biligo-live-ws/releases)

运行参数(非必要)

```bash
./biligo-live-ws 端口
```

端口: 不填则 8080

## 鸣谢

[bili-go](https://github.com/iyear/biligo-live) 作者
