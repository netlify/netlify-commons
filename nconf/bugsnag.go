package nconf

import (
	"github.com/bugsnag/bugsnag-go/v2"
)

type BugSnagConfig struct {
	Environment    string
	APIKey         string `envconfig:"api_key" json:"api_key" yaml:"api_key"`
	ProjectPackage string `envconfig:"project_package" json:"project_package" yaml:"project_package"`
	CDNHostName    string `envconfig:"node_name" json:"node_name" yaml:"node_name"`
}

func SetupBugSnag(config *BugSnagConfig, version string) error {
	if config == nil || config.APIKey == "" {
		return nil
	}

	projectPackages := make([]string, 0, 2)
	projectPackages = append(projectPackages, "main")
	if config.ProjectPackage != "" {
		projectPackages = append(projectPackages, config.ProjectPackage)
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:          config.APIKey,
		ReleaseStage:    config.Environment,
		Hostname:        config.CDNHostName,
		AppVersion:      version,
		ProjectPackages: projectPackages,
		PanicHandler:    func() {}, // this is to disable panic handling. The lib was forking and restarting the process (causing races)
	})

	return nil
}
