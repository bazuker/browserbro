# BrowserBro 😎
BrowserBro is a simple HTTP server and a collection of plugins that are useful for web scraping, testing and many other things.
At its core, BrowserBro is using [Rod](https://github.com/go-rod/rod), a high-level Go browser automation library.

The goal of BrowserBro is to provide a simple yet powerful way to run browser automation scripts, without the need to manage the browser lifecycle.

## Plugins ⚙️
BrowserBro comes with a collection of plugins that can be used to perform various tasks.
The plugins are loaded dynamically, so you can easily extend the functionality of BrowserBro by adding your own plugins.

Plugins are available as HTTP endpoints and can be accessed by sending a POST request to the server.
You can supply the plugin with the necessary parameters in the request body.
```
POST /api/v1/plugins/{plugin-name}
```

#### Example
Look how simple it is to scrape google search results with BrowserBro 🔍
```bash
curl -X POST -d '{"query":"latest Golang news"}' http://localhost:10001/api/v1/plugins/googlesearch
``` 
Response:
```json
{
  "googlesearch": {
    "all": [
      {
        "description": "A weekly newsletter about the Go programming language ... Check out our latest issue for a sample. ... Our privacy, anti-spam, and GDPR policies. We take these ...",
        "link": "https://golangweekly.com/",
        "title": "Golang Weekly"
      },
      ...
    ]
  }
}
```

### Available plugins

To get a list of all available plugins:
```
GET /api/v1/plugins
```

#### Google Search
Name: `googlesearch`

Parameters:
- `query` [String] - The search query
- `type` [Strings array] - The type of search results to return. Possible values: `all`, `videos`

#### Screenshot
Name: `screenshot`

Parameters:
- `urls` [Strings array] - The URLs is list of links to the pages to take screenshots of. Can be a single URL or a list of URLs.
- `waitStable` [Boolean] - Wait until the page is stable before taking a screenshot. Default: `true`

## Files 📁
BrowserBro can also serve static files generated or downloaded by the plugins.
The files are available at the following URL:
```bash
GET /api/v1/files/{fileID}
```
Files can also be deleted by sending a DELETE request to the same URL.
```bash
DELETE /api/v1/files/{fileID}
```

#### Example
When you use the `screenshot` plugin, the plugin will save the screenshots as files and return the file IDs in the response.
```json
{
  "screenshot": {
    "files": [
      "Shu2vLZm.screenshot.png"
    ]
  }
}
```
You can then access the files by sending a GET request to the file URL.
```bash
curl http://localhost:10001/api/v1/files/Shu2vLZm.screenshot.png
```


## Health Check
To check if the server is running, you can send a GET request to the health check endpoint.
```
GET /api/v1/health
```