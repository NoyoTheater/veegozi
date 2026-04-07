# veegozi

A Go module to work with the [Veezi API](https://api.us.veezi.com/).

## Features

- Full coverage of Veezi API endpoints
- Strongly typed data structures
- Caching

## Installation

```bash
go get github.com/NoyoTheater/veegozi
```

## Usage

```go
client := NewClient("api.us.veezi.com", "your_api_key", WithDefaultCaching())
sessions, err := client.GetSessions(context.Background())
```
