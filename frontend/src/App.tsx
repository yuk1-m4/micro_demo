import { useEffect, useState } from 'react'
import { describeHealth } from './health'
import './App.css'

function App() {
  const [status, setStatus] = useState('checking...')

  useEffect(() => {
    fetch('/api/health')
      .then((res) => (res.ok ? res.json() : Promise.reject(res.status)))
      .then((body) => setStatus(describeHealth(body)))
      .catch(() => setStatus('backend unreachable'))
  }, [])

  return (
    <>
      <h1>micro_demo</h1>
      <p>backend: {status}</p>
    </>
  )
}

export default App
