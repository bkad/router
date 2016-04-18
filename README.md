# Deis Router - Experimental Caddy fork

Disclaimer: This is not supported or endorsed by Deis (Engine Yard). Fun times using the awesome code that is in Deis router - nothing more at this point.

# About

This is a proof-of-concept fork of the [Deis Router][deisrouter] which uses [Caddy Server][caddy] instead on NGINX. This has only been tested outside of the [Deis platform][deis], it is in no way ready for use as a drop in replacement in [Deis Workflow][deisworkflow]. It has only been tested outside of the Deis platform on a bare kubernetes cluster. There are fundamental issues with using this as a drop in replacement, so don't expect that to work.

The goal is to utilize the automatic [Let's Encrypt][letsencrypt] integration provided by Caddy to register SSL certs for all domains added to the router. This router does not yet store these certs in any persistent storage, it is not nearly as configurable as the current Deis nginx router, it has not at all been tested at scale or within the Deis platform, it's bound by the Let's Encrypt rate limits per domain, and is not for serious use at this time.

All of the configurability and cert handling in the Deis Router V2 was removed to simplify the proof of concept. Most of this code will be brought back to allow the root domain to provide a wildcard cert and for individual domains to provide their own certs and allow more configuration of the caddy server.

Please refer to the [Deis Router V2 documentation][deisrouter]. Below is the minimal configurations currently supported by this caddy fork of the router.

### Configurations via Kubernetes Manifests

| Component | Resource Type | Label/Annotation | Key | Description |
|-----------|---------------|------------|---------------|-------------|
| deis-router | RC | annotation| router.deis.io/caddy.platformDomain | This defines the router's platform domain.  Any domains added to a routable application _not_ containing the `.` character will be assumed to be subdomains of this platform domain.  Thus, for example, a platform domain of `example.com` coupled with a routable app counting `foo` among its domains will result in router configuration that routes traffic for `foo.example.com` to that application. |
| routable application | service | label | router.deis.io/routable | Only services that have a `routable` value of `"true"` will be tracked. |
| routable application | service | annotation | router.deis.io/domains | Comma-delimited list of domains for which traffic should be routed to the application.  These may be fully qualified (e.g. `foo.example.com`) or, if not containing any `.` character, will be considered subdomains of the router's domain, if that is defined. |

#### Configurations by example

##### router replication controller:

```
apiVersion: v1
kind: ReplicationController
metadata:
  name: deis-router
  namespace: deis
  # ...
  annotations:
    router.deis.io/caddy.platformDomain: example.com
    router.deis.io/tlsEmail: dr@who.com
# ...
```

##### routable service:

```
apiVersion: v1
kind: Service
metadata:
  name: foo
  labels:
  	router.deis.io/routable: "true"
  namespace: examples
  # ...
  annotations:
    router.deis.io/domains: foo,bar,www.foobar.com
    router.deis.io/tlsEmail: dr@who.com
    # router.deis.io/tls: off      (only if you have to)
# ...
```

## License

The original Deis Router is Copyright Engine Yard, Inc.

Caddy Server comes from Matt Holt.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

[deis]: https://deis.com
[deisrouter]: https://github.com/deis/router/
[deisworkflow]: https://github.com/deis/workflow
[caddy]: https://caddyserver.com
[letsencrypt]: https://letsencrypt.org/
