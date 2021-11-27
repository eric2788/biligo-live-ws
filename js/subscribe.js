const SUBSCRIBE_URL = `http://${API_HOST}/subscribe`


const api = axios.create({
    baseURL: SUBSCRIBE_URL,
    timeout: 5000,
    headers: {
        'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8'
    }
})

async function getSubscribtions(){
    const response = await api.get('')
    if (response.status !== 200){
        throw new Error(response.statusText)
    }
    return response.data
}


async function subscribes(list){
    const response = await api.post('', { subscribes: list })
    if (response.status !== 200){
        throw new Error(response.statusText)
    }
    return response.data
}


async function clearSubscribe(){
    const response = await api.delete('')
    if (response.status !== 200){
        throw new Error(response.statusText)
    }
    return response.data
}

// fake service

async function sleep(ms) {
    return new Promise((res, ) => setTimeout(res, ms))
}

let subscribeList = [
    123456789,
    987654321,
    123456,
    45454545,
    114514,
    1919810
]

async function getSubscribtionsFake(){
    await sleep(3000)
    return subscribeList
}


const invalids = [ 123456, 114514, 1919810 ]

async function subscribesFake(list){
    subscribeList = subscribeList.filter(room => !invalids.includes(room))
    await sleep(2500)
    return subscribeList
}


async function clearSubscribeFake(){
    subscribeList = []
    await sleep(2500)
    return {}
}