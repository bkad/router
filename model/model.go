package model

import (
	"fmt"
	"log"

	"github.com/drud/router/utils"
	modelerUtility "github.com/drud/router/utils/modeler"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

const (
	prefix               string = "router.deis.io"
	modelerFieldTag      string = "key"
	modelerConstraintTag string = "constraint"
)

var (
	namespace        = utils.GetOpt("POD_NAMESPACE", "default")
	modeler          = modelerUtility.NewModeler(prefix, modelerFieldTag, modelerConstraintTag, true)
	servicesSelector labels.Selector
)

func init() {
	var err error
	servicesSelector, err = labels.Parse(fmt.Sprintf("%s/routable==true", prefix))
	if err != nil {
		log.Fatal(err)
	}
}

// RouterConfig is the primary type used to encapsulate all router configuration.
type RouterConfig struct {
	PlatformDomain string `key:"platformDomain" constraint:"(?i)^([a-z0-9]+(-[a-z0-9]+)*\\.)+[a-z]{2,}$"`
	AppConfigs     []*AppConfig
	BuilderConfig  *BuilderConfig
	TLS            string `key:"tls" constraint:"^(off)$`
	TLSEmail       string `key:"tlsEmail"`
}

func newRouterConfig() *RouterConfig {
	return &RouterConfig{
		TLS:      "",
		TLSEmail: "",
	}
}

// AppConfig encapsulates the configuration for all routes to a single back end.
type AppConfig struct {
	Name          string
	Domains       []string `key:"domains" constraint:"(?i)^((([a-z0-9]+(-[a-z0-9]+)*)|((\\*\\.)?[a-z0-9]+(-[a-z0-9]+)*\\.)+[a-z]{2,})(\\s*,\\s*)?)+$"`
	TLS           string   `key:"tls" constraint:"^(off)$`
	TLSEmail      string   `key:"tlsEmail"`
	BasicAuthPath string   `key:"basicAuthPath"`
	BasicAuthUser string   `key:"basicAuthUser"`
	BasicAuthPass string   `key:"basicAuthPass"`
	ServiceIP     string
	Available     bool
}

func newAppConfig(routerConfig *RouterConfig) *AppConfig {
	return &AppConfig{
		TLS:           "",
		TLSEmail:      "",
		BasicAuthPath: "/",
		BasicAuthUser: "",
		BasicAuthPass: "",
	}
}

// BuilderConfig encapsulates the configuration of the deis-builder-- if it's in use.
type BuilderConfig struct {
	ServiceIP string
}

func newBuilderConfig() *BuilderConfig {
	return &BuilderConfig{}
}

// Build creates a RouterConfig configuration object by querying the k8s API for
// relevant metadata concerning itself and all routable services.
func Build(kubeClient *client.Client) (*RouterConfig, error) {
	// Get all relevant information from k8s:
	//   deis-router rc
	//   All services with label "routable=true"
	//   deis-builder service, if it exists
	// These are used to construct a model...
	routerRC, err := getRC(kubeClient)
	if err != nil {
		return nil, err
	}
	appServices, err := getAppServices(kubeClient)
	if err != nil {
		return nil, err
	}
	// builderService might be nil if it's not found and that's ok.
	builderService, err := getBuilderService(kubeClient)
	if err != nil {
		return nil, err
	}
	// Build the model...
	routerConfig, err := build(kubeClient, routerRC, appServices, builderService)
	if err != nil {
		return nil, err
	}
	return routerConfig, nil
}

func getRC(kubeClient *client.Client) (*api.ReplicationController, error) {
	rcClient := kubeClient.ReplicationControllers(namespace)
	rc, err := rcClient.Get("deis-router")
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func getAppServices(kubeClient *client.Client) (*api.ServiceList, error) {
	serviceClient := kubeClient.Services(api.NamespaceAll)
	services, err := serviceClient.List(servicesSelector)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// getBuilderService will return the service named "deis-builder" from the same namespace as
// the router, but will return nil (without error) if no such service exists.
func getBuilderService(kubeClient *client.Client) (*api.Service, error) {
	serviceClient := kubeClient.Services(namespace)
	service, err := serviceClient.Get("deis-builder")
	if err != nil {
		statusErr, ok := err.(*errors.StatusError)
		// If the issue is just that no deis-builder was found, that's ok.
		if ok && statusErr.Status().Code == 404 {
			// We'll just return nil instead of a found *api.Service.
			return nil, nil
		}
		return nil, err
	}
	return service, nil
}

func build(kubeClient *client.Client, routerRC *api.ReplicationController, appServices *api.ServiceList, builderService *api.Service) (*RouterConfig, error) {
	routerConfig, err := buildRouterConfig(routerRC)
	if err != nil {
		return nil, err
	}
	for _, appService := range appServices.Items {
		appConfig, err := buildAppConfig(kubeClient, appService, routerConfig)
		if err != nil {
			return nil, err
		}
		if appConfig != nil {
			routerConfig.AppConfigs = append(routerConfig.AppConfigs, appConfig)
		}
	}
	if builderService != nil {
		builderConfig, err := buildBuilderConfig(builderService)
		if err != nil {
			return nil, err
		}
		if builderConfig != nil {
			routerConfig.BuilderConfig = builderConfig
		}
	}
	return routerConfig, nil
}

func buildRouterConfig(rc *api.ReplicationController) (*RouterConfig, error) {
	routerConfig := newRouterConfig()
	err := modeler.MapToModel(rc.Annotations, "caddy", routerConfig)
	if err != nil {
		return nil, err
	}
	return routerConfig, nil
}

func buildAppConfig(kubeClient *client.Client, service api.Service, routerConfig *RouterConfig) (*AppConfig, error) {
	appConfig := newAppConfig(routerConfig)
	appConfig.Name = service.Labels["app"]
	// If we didn't get the app name from the app label, fall back to inferring the app name from
	// the service's own name.
	if appConfig.Name == "" {
		appConfig.Name = service.Name
	}
	err := modeler.MapToModel(service.Annotations, "", appConfig)
	if err != nil {
		return nil, err
	}
	// If no domains are found, we don't have the information we need to build routes
	// to this application.  Abort.
	if len(appConfig.Domains) == 0 {
		return nil, nil
	}
	appConfig.ServiceIP = service.Spec.ClusterIP
	endpointsClient := kubeClient.Endpoints(service.Namespace)
	endpoints, err := endpointsClient.Get(service.Name)
	if err != nil {
		return nil, err
	}
	appConfig.Available = len(endpoints.Subsets) > 0 && len(endpoints.Subsets[0].Addresses) > 0
	return appConfig, nil
}

func buildBuilderConfig(service *api.Service) (*BuilderConfig, error) {
	builderConfig := newBuilderConfig()
	builderConfig.ServiceIP = service.Spec.ClusterIP
	err := modeler.MapToModel(service.Annotations, "caddy", builderConfig)
	if err != nil {
		return nil, err
	}
	return builderConfig, nil
}
