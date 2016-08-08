# ernest-config-client

This library is intended to be a client to build services based on ernest-config service.

## Installation

```
go get -u github.com/ernestio/ernest-config-client
```

## Usage

```
import "github.com/ernestio/ernest-config-client"

func main() {
  c := ernest_config_client.NewConfig("nats://127.0.0.1:4222")

  # Get a redis client
  r := c.Redis()

  # Get a postgres client
  p := c.Postgres()

  # get the nats client
  n := c.Nats()
}
```

## Contributing

Please read through our
[contributing guidelines](CONTRIBUTING.md).
Included are directions for opening issues, coding standards, and notes on
development.

Moreover, if your pull request contains patches or features, you must include
relevant unit tests.

## Versioning

For transparency into our release cycle and in striving to maintain backward
compatibility, this project is maintained under [the Semantic Versioning guidelines](http://semver.org/).

## Copyright and License

Code and documentation copyright since 2015 r3labs.io authors.

Code released under
[the Mozilla Public License Version 2.0](LICENSE).
