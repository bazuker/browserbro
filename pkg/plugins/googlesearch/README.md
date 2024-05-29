# Google search 🔍

Name: `googlesearch`

Parameters:
- `query` [String] - The search query
- `type` [Strings array] - The type of search results to return. Possible values: `all`, `videos`


Response format:
```json
{
  "googlesearch": {
    "all": [
      {
        "description": "A weekly newsletter about the Go programming language ...",
        "link": "https://golangweekly.com/",
        "title": "Golang Weekly"
      },
      ...
    ]
  }
}
```