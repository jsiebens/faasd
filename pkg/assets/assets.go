package assets

import _ "embed"

//go:embed docker-compose.yaml
var DockerCompose []byte

//go:embed prometheus.yml
var PrometheusConfig []byte

//go:embed resolv.conf
var ResolvConf []byte

//go:embed faasd.service
var FaasdService []byte

//go:embed faasd-provider.service
var FaasdProviderService []byte
