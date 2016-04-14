package caddy

import (
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/drud/router/model"
)

const (
	confTemplate = `# Automatically generated Caddyfile
0.0.0.0 {
	root /opt/router/default
}
{{ $routerConfig := . }}
{{ range $appConfig := $routerConfig.AppConfigs }}{{ range $domain := $appConfig.Domains }}{{ if $appConfig.Available }}
{{ if contains "." $domain }}{{ $domain }}{{ else if ne $routerConfig.PlatformDomain "" }}{{ $domain }}.{{ $routerConfig.PlatformDomain }}{{ else }}{{ $domain }}{{ end }} {
	proxy / {{$appConfig.ServiceIP}}:80
}
{{ end }}{{ end }}{{ end }}
`
)

// WriteConfig dynamically produces valid caddy configuration by combining a Router configuration
// object with a data-driven template.
func WriteConfig(routerConfig *model.RouterConfig, filePath string) error {
	tmpl, err := template.New("caddy").Funcs(sprig.TxtFuncMap()).Parse(confTemplate)
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	err = tmpl.Execute(file, routerConfig)
	if err != nil {
		return err
	}
	return nil
}