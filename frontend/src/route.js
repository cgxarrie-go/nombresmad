import { useEffect, useState } from 'react'

const listeners = new Set()

export function navigate(path) {
  if (window.location.pathname !== path) {
    window.history.pushState(null, '', path)
  }
  listeners.forEach((listener) => listener(window.location.pathname))
}

export function usePath() {
  const [path, setPath] = useState(window.location.pathname)
  useEffect(() => {
    const sync = () => setPath(window.location.pathname)
    listeners.add(sync)
    window.addEventListener('popstate', sync)
    return () => {
      listeners.delete(sync)
      window.removeEventListener('popstate', sync)
    }
  }, [])
  return path
}
