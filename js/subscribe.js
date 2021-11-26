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