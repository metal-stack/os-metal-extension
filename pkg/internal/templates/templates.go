package templates

import _ "embed"

// CloudInitTemplate contains the content of the cloud-init.sh.template file
//go:embed cloud-init.sh.template
var CloudInitTemplate string
