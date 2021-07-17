# Http2MQTT

Http2MQTT is a one-way HTTP to MQTT publisher that allows for publishing messages to arbitrary topics on a broker.

## Installing Http2MQTT

```
go get github.com/billyoverton/http2mqtt
```

## Using Http2MQTT

```
A HTTP to MQTT Broker Bridge

Usage:
  http2mqtt [flags]

Flags:
  -b, --broker string     MQTT broker address (default "localhost")
  -q, --brokerport int    port to listen for web connections (default 1883)
  -c, --config string     config file (default is $HOME/.http2mqtt.yaml)
  -h, --help              help for http2mqtt
  -P, --password string   broker password
  -p, --port int          port to listen for web connections (default 8080)
  -u, --username string   broker username
```

Http2MQTT uses [Viper](https://github.com/spf13/viper) for it's configuration, so all flags can be set using a configuration file or through environment variables.


## License
[MIT](LICENSE)
