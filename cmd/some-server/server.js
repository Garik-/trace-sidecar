const http = require('node:http');
const url = require("url");

const hostname = 'localhost';
const port = 8000;

const server = http.createServer((req, res) => {
    res.statusCode = 404;

    // Parse the request url
    const { pathname } = url.parse(req.url)
    if (pathname === "/ok") {
        res.statusCode = 200;
    }

    res.end()
});

server.listen(port, hostname, () => {
    console.log(`Server running at http://${hostname}:${port}/`);
}); 