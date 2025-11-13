# blog.reyel.dev

My personal blog

## Development

### Prerequisites

- [Go](https://golang.org/) installed.

### Installation

Install the Go dependencies:

```sh
go get
```

### Commands

**Generate**

To convert the markdown files in `blog/` to HTML in `dist/`:

```sh
go run main.go generate
```

**Serve**

To serve the generated HTML files from `dist/` on `http://localhost:8080`:

```sh
go run main.go serve
```

## TODO

- Styling
