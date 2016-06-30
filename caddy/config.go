package caddy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/drud/router/model"
)

const (
	confTemplate = `# Automatically generated Caddyfile
0.0.0.0 {
	header / HTTP/1.0 "404 Not Found"
}

{{ $routerConfig := . }}

{{ range $appConfig := $routerConfig.AppConfigs }}{{ range $domain := $appConfig.Domains }}{{ if $appConfig.Available }}
{{ if contains "." $domain }}{{ $domain }}{{ else if ne $routerConfig.PlatformDomain "" }}{{ $domain }}.{{ $routerConfig.PlatformDomain }}{{ else }}{{ $domain }}{{ end }} {
    proxy / {{$appConfig.ServiceIP}}:80 {
        proxy_header Host {host}
        proxy_header X-Forwarded-Proto {scheme}
    }
    {{ if eq $routerConfig.TLS "off" }}
    tls off
    {{ else if eq $appConfig.TLS "off" }}
    tls off
    {{ else }}
        {{ if not contains "." $domain and ne $routerConfig.PlatformDomain "" and $routerConfig.PlatformCertificate }}
    tls /opt/router/ssl/platform.crt /opt/router/ssl/platform.key
        {{ else if $appConfig.TLSEmail }}
    tls {{ $appConfig.TLSEmail }}
        {{ else if $routerConfig.TLSEmail }}
    tls {{ $routerConfig.TLSEmail }}
        {{ end }}
    {{ end }}
    {{ if ne $appConfig.BasicAuthPath "" }}{{ if ne $appConfig.BasicAuthUser "" }}{{ if ne $appConfig.BasicAuthPass "" }}
    basicauth {{ $appConfig.BasicAuthPath }} {{ $appConfig.BasicAuthUser }} {{ $appConfig.BasicAuthPass }}
    {{ end }}{{ end }}{{ end }}
}
{{ end }}{{ end }}{{ end }}
`
)

func WriteCerts(routerConfig *model.RouterConfig, sslPath string) error {
	// Start by deleting all certs and their corresponding keys. This will ensure certs we no longer
	// need are deleted. Certs that are still needed will simply be re-written.
	allCertsGlob, err := filepath.Glob(filepath.Join(sslPath, "*.crt"))
	if err != nil {
		return err
	}
	allKeysGlob, err := filepath.Glob(filepath.Join(sslPath, "*.key"))
	if err != nil {
		return err
	}
	for _, cert := range allCertsGlob {
		if err := os.Remove(cert); err != nil {
			return err
		}
	}
	for _, key := range allKeysGlob {
		if err := os.Remove(key); err != nil {
			return err
		}
	}
	if routerConfig.PlatformCertificate != nil {
		err = writeCert("platform", routerConfig.PlatformCertificate, sslPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeCert(context string, certificate *model.Certificate, sslPath string) error {
	certPath := filepath.Join(sslPath, fmt.Sprintf("%s.crt", context))
	keyPath := filepath.Join(sslPath, fmt.Sprintf("%s.key", context))
	err := ioutil.WriteFile(certPath, []byte(certificate.Cert), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(keyPath, []byte(certificate.Key), 0600)
	if err != nil {
		return err
	}
	return nil
}

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
