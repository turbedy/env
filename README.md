# env
[![Release](https://img.shields.io/github/v/release/turbedy/env?logo=github)](https://github.com/turbedy/env/releases/latest)
[![CI](https://img.shields.io/github/actions/workflow/status/turbedy/env/ci.yaml?logo=github&label=ci)](https://github.com/turbedy/env/actions/workflows/ci.yaml)
[![Coverage](https://img.shields.io/codecov/c/github/turbedy/env?logo=codecov&token=09RNPBR0K0)](https://codecov.io/github/turbedy/env)

A Go package for mapping environment variables to structs.

## Getting started
Install the package:
```bash
go get github.com/turbedy/env/v2
```

Define a struct:
```go
type Address struct {
    Host string
    Port int
}

type Config struct {
    HTTP Address
    IPv6Enabled bool
}
```

Set environment variables:
```bash
export HTTP_HOST=::
export HTTP_PORT=80
export IPV6_ENABLED=true
```

Map environment variables:
```go
var cfg Config
if err := env.Decode(&cfg); err != nil {
    log.Fatal(err)
}
```

You can see the full documentation at [pkg.go.dev](https://pkg.go.dev/github.com/turbedy/env/v2).

## License
This project is licensed under the [BSD-3-Clause](LICENSE).
