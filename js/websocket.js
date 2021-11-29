
const WEBSOCKET_URL = `ws${SCHEMA}://${API_HOST}/ws?id=${IDENTIFIER}`

console.log(`websocket url is: ${WEBSOCKET_URL}`)

// == real

let ws;

async function connectReal(callback){
    await validate() // validate before open websocket
    return new Promise((res, rej) => {
        ws = new WebSocket(WEBSOCKET_URL)
        ws.onopen = res
        ws.onmessage = function (e) {
            try {
                callback(JSON.parse(e.data))
            }catch(err){
                console.warn(`接收訊息時出現錯誤: ${err}`)
            }
        }
        ws.onerror = function(e){
            console.warn(`WebSocket Error`)
            console.warn(e)
            rej(e)
        }
    })

}

async function disconnectReal(){
    if (!ws) return
    await ws.close()
    return new Promise((res, ) => ws.onclose = res)
}


// === fake 

let timer = -1

async function connectFake(callback){
    await sleep(2000)
    timer = setInterval(() => {
        callback(Math.random() * 12345)
    }, 100)
}

async function disconnectFake(){
    clearInterval(timer)
    await sleep(3000)
}

// == combine


async function connect(callback){
    return connectReal(callback)
}


async function disconnect(){
    return disconnectReal()
}
