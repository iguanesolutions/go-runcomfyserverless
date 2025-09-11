package runcomfyserverless

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-cleanhttp"
)

type Deployment struct {
	auth    string
	baseURL *url.URL
	http    *http.Client
}

type Config struct {
	UserAPIToken     string
	DeploymentID     string
	CustomHTTPClient *http.Client // can be left empty
}

func LinkToDeployment(conf Config) (deploymentClient *Deployment) {
	if conf.CustomHTTPClient == nil {
		conf.CustomHTTPClient = cleanhttp.DefaultPooledClient()
	}
	return &Deployment{
		auth:    fmt.Sprintf("Bearer %s", conf.UserAPIToken),
		baseURL: baseURL.JoinPath("prod/v1/deployments", conf.DeploymentID),
		http:    conf.CustomHTTPClient,
	}
}
