const { createProxyMiddleware } = require('http-proxy-middleware');

// proxy URI
const URI = process.env.REVERSE_PROXY_URI || 'http://localhost:8080';

module.exports = (app) => {
  app.use(createProxyMiddleware('/api', { target: URI, ws: true }));
};
