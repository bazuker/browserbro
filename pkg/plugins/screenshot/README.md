# Screenshot 📸

Name: `screenshot`

Parameters:
- `urls` [Strings array] - The URLs is list of links to the pages to take screenshots of. Can be a single URL or a list of URLs.
- `waitStable` [Boolean] - Wait until the page is stable before taking a screenshot. Default: `true`

Response format:
```json
{
  "screenshot": {
    "files": [
      "Shu2vLZm.screenshot.png",
      "Az42KhY9.screenshot.png"
    ]
  }
}
```