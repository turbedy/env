# env
[![Release](https://img.shields.io/github/v/release/turbedy/env?logo=github)](https://github.com/turbedy/env/releases/latest)
[![CI](https://img.shields.io/github/actions/workflow/status/turbedy/env/ci.yaml?logo=github&label=ci)](https://github.com/turbedy/env/actions/workflows/ci.yaml)
[![Coverage](https://img.shields.io/codecov/c/github/turbedy/env?logo=codecov&token=09RNPBR0K0)](https://codecov.io/github/turbedy/env)

A Go package for mapping environment variables to structs.

## Usage
Install the package:
```bash
go get github.com/turbedy/env
```

Define nested structs:
```go
type Address struct {
    Host string
    Port int
}

type Server struct {
    Address
    Headers map[string]string
}

type Config struct {
    HTTP Server
    IPv6Enabled bool
}
```

Set environment variables:
```bash
export HTTP_HOST=::1
export HTTP_PORT=80
export HTTP_HEADERS=Content-Type:application/json,Cache-Control:no-cache
export IPV6_ENABLED=true
```

Map environment variables to exported fields:
```go
var cfg Config
if err := env.Decode(&cfg); err != nil {
    log.Fatal(err)
}
```

You can see the full documentation at [pkg.go.dev](https://pkg.go.dev/github.com/turbedy/env).

## License
This project is licensed under the [BSD-3-Clause](LICENSE).
