
const WEBSOCKET_URL = `ws://${API_HOST}/ws`

console.log(`websocket url is: ${WEBSOCKET_URL}`)

function connect(){

}


// === fake 

let timer = -1

async function fakeConnect(callback){
    await sleep(2000)
    timer = setInterval(() => {
        callback(Math.random() * 12345)
    }, 100)
}

async function fakeDisconnect(){
    clearInterval(timer)
    await sleep(3000)
}
