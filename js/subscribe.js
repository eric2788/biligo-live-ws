const SUBSCRIBE_URL = `http${SCHEMA}://${API_HOST}/subscribe`

const VALIDATE_URL = `http${SCHEMA}://${API_HOST}/validate`

async function validate(){
    const res = await axios.post(VALIDATE_URL)
    if (res.status !== 200){
        throw new Error(res.statusText)
    }
}


const api = axios.create({
    baseURL: SUBSCRIBE_URL,
    timeout: 5000,
    headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
    }
})


// == real

async function getSubscribtionsReal(){
    const response = await api.get('')
    if (response.status !== 200){
        throw new Error(response.statusText)
    }
    return response.data
}


async function subscribesReal(list){
    const form = new FormData()
    for (const room of list){
        form.append('subscribes', room)
    }
    const response = await api.post('', form)
    if (response.status !== 200){
        throw new Error(response.statusText)
    }
    return response.data
}


async function clearSubscribeReal(){
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
    subscribeList = list.filter(room => !invalids.includes(room))
    await sleep(2500)
    return subscribeList
}


async function clearSubscribeFake(){
    subscribeList = []
    await sleep(2500)
    return {}
}


// combine

async function getSubscribtions(){
    return getSubscribtionsReal()
}


async function subscribes(list){
    return subscribesReal(list)
}


async function clearSubscribe(){
    return clearSubscribeReal()
}