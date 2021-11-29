function BLiveDataFomatter(props) {

    const msg = props.data
    const { name: liveName } = msg.live_info


    let content = msg.content //照計應該是 object
    
    if (typeof content === 'string'){ // 但如果因為錯誤變成了 string
        try {
            content = JSON.parse(msg.content)
        } catch (err) {
            console.warn(`轉換 json 時出現錯誤: ${err.message}`)
            console.warn(err)
            console.warn(msg.content)
            return
        }
    }

    if (msg.command === 'HEARTBEAT_REPLY'){
        return `[${liveName}直播间] 房间人气值: ${content.popularity}`
    }   

    if (content.cmd === 'DANMU_MSG') {
        const [, danmaku, [uid, uname]] = content.info
        return `[${liveName}直播间] ${uname}: ${danmaku}`
    } else if (content.cmd === 'SEND_GIFT'){
        const data = content.data
        return `[${liveName}直播间] ${data.uname} ${data.action}了 ${data.giftName}x${data.num}`
    }else if (content.cmd === 'INTERACT_WORD'){
        const { uname } = content.data
        return `[${liveName}直播间] ${uname} 进入了直播间`
    }else if (content.cmd === 'LIVE'){
        return `${msg.name} 开播了`
    } else if (content.cmd === 'SUPER_CHAT_MESSAGE'){
        const { price, message, user_info } = content.data
        return `[${liveName}直播间] ${user_info.uname} [￥${price}]: ${message}`
    }else{
        return null
    }
}