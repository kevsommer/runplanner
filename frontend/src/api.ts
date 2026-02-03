import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.DEV ? '/api' : 'http://localhost:8080/api',
  withCredentials: true, // send cookie on XHR/fetch
})
