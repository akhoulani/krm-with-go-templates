// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const SERVICE_TEMPLATE string = `
apiVersion: v1
kind: Service
metadata:
 name: {{ .Metadata.Name }}
spec:
 selector:
   app: {{ .Metadata.Name }}
 ports:
 - port: {{ .Spec.Port }}
   targetPort: {{ .Spec.Port }}`

var _ fn.Runner = &YourFunction{}

type Metadata struct {
	Name string `yaml:"name"`
}

type AppSpec struct {
	Image string `yaml:"image"`
	Port  int32  `yaml:"port"`
	Size  string `yaml:"size,omitempty"`
}

type App struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     AppSpec  `yaml:"spec"`
}

// TODO: Change to your functionConfig "Kind" name.
type YourFunction struct {
	FnConfigBool bool
	FnConfigInt  int
	FnConfigFoo  string
}

// Run is the main function logic.
// `items` is parsed from the STDIN "ResourceList.Items".
// `functionConfig` is from the STDIN "ResourceList.FunctionConfig". The value has been assigned to the r attributes
// `results` is the "ResourceList.Results" that you can write result info to.
func (r *YourFunction) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	for _, kubeobject := range items {
		if kubeobject.IsGVK("apps", "v1", "Deployment") {
			kubeobject.SetAnnotation("config.kubernetes.io/managed-by", "amer-kpt")
		}
	}
	// This result message will be displayed in the function evaluation time.
	*results = append(*results, fn.GeneralResult("Add config.kubernetes.io/managed/by=amer-kpt to all `Deployment` resources", fn.Info))
	return true
}

func filterAppFromResources(items []*yaml.RNode) ([]*yaml.RNode, error) {
	var newNodes []*yaml.RNode
	for i := range items {
		meta, err := items[i].GetMeta()
		if err != nil {
			return nil, err
		}
		// remove resources with the kind App from the resource list
		if meta.Kind == "App" && meta.APIVersion == "app.innoq.com/v1" {
			continue
		}
		newNodes = append(newNodes, items[i])
	}
	items = newNodes
	return items, nil
}

func main() {
	config := &App{}
	fn := framework.TemplateProcessor{
		TemplateData:       config,
		PostProcessFilters: []kio.Filter{kio.FilterFunc(filterAppFromResources)},
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: parser.TemplateStrings(DEPLOYMENT_TEMPLATE, SERVICE_TEMPLATE),
		}},
	}
	/*
	   cmd := command.Build(fn, command.StandaloneDisabled, false)
	   command.AddGenerateDockerfile(cmd)

	   	if err := cmd.Execute(); err != nil {
	   		os.Exit(1)
	   	}

	   	runner := fn.WithContext(context.Background(), &YourFunction{})
	   	if err := fn.AsMain(runner); err != nil {
	   		os.Exit(1)
	   	}
	*/
}
