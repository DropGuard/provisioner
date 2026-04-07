package provisioner

import _ "embed"

//go:embed config.yaml
var EmbeddedConfig []byte
