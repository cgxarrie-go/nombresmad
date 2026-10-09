const fs = require('fs')
const http = require('http')
const https = require('https')
const path = require('path')

const buildRoot = path.join(__dirname, 'build')
const backend = new URL(process.env.API_PROXY_TARGET || 'http://localhost:8080')
const contentTypes = {
  '.css': 'text/css; charset=utf-8',
  '.html': 'text/html; charset=utf-8',
  '.ico': 'image/x-icon',
  '.js': 'application/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.png': 'image/png',
  '.svg': 'image/svg+xml',
  '.txt': 'text/plain; charset=utf-8',
  '.webp': 'image/webp',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
}

function proxyRequest(req, res) {
  const transport = backend.protocol === 'https:' ? https : http
  const request = transport.request(
    {
      protocol: backend.protocol,
      hostname: backend.hostname,
      port: backend.port || undefined,
      method: req.method,
      path: req.url,
      headers: { ...req.headers, host: backend.host },
    },
    (response) => {
      res.writeHead(response.statusCode, response.headers)
      response.pipe(res)
    }
  )

  request.on('error', (error) => {
    console.error('Backend proxy error:', error.message)
    if (!res.headersSent) {
      res.writeHead(502, { 'Content-Type': 'text/plain; charset=utf-8' })
    }
    res.end('Backend unavailable')
  })
  req.pipe(request)
}

function serveFile(req, res, filePath) {
  fs.stat(filePath, (error, stats) => {
    const resolvedPath = !error && stats.isFile() ? filePath : path.join(buildRoot, 'index.html')
    fs.readFile(resolvedPath, (readError, content) => {
      if (readError) {
        res.writeHead(404)
        res.end('Not found')
        return
      }
      res.writeHead(200, {
        'Content-Type': contentTypes[path.extname(resolvedPath)] || 'application/octet-stream',
        'Content-Length': content.length,
      })
      if (req.method === 'HEAD') res.end()
      else res.end(content)
    })
  })
}

http
  .createServer((req, res) => {
    const pathname = new URL(req.url, 'http://localhost').pathname
    if (pathname === '/api' || pathname.startsWith('/api/')) {
      proxyRequest(req, res)
      return
    }
    if (req.method !== 'GET' && req.method !== 'HEAD') {
      res.writeHead(405, { Allow: 'GET, HEAD' })
      res.end()
      return
    }

    let decodedPath
    try {
      decodedPath = decodeURIComponent(pathname)
    } catch {
      res.writeHead(400)
      res.end('Bad request')
      return
    }
    const filePath = path.resolve(buildRoot, `.${decodedPath}`)
    if (!filePath.startsWith(`${buildRoot}${path.sep}`) && filePath !== buildRoot) {
      res.writeHead(403)
      res.end('Forbidden')
      return
    }
    serveFile(req, res, filePath)
  })
  .listen(Number(process.env.PORT) || 3000, '0.0.0.0', () => {
    console.log(`Frontend listening on port ${Number(process.env.PORT) || 3000}`)
    console.log(`Proxying /api to ${backend.origin}`)
  })