module github.com/brigadecore/brigade

go 1.14

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.0.1+incompatible
	k8s.io/client-go => k8s.io/client-go v0.18.2
)

require (
	cloud.google.com/go v0.53.0 // indirect
	github.com/Azure/go-autorest/autorest v0.10.0 // indirect
	github.com/Masterminds/goutils v1.1.0
	github.com/Masterminds/kitt v0.0.0-20160203155249-7e843d5f21a5
	github.com/bacongobbler/browser v1.1.0
	github.com/cloudevents/sdk-go v0.0.0-20190102195109-feec6e002535
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/emicklei/go-restful v2.11.2+incompatible
	github.com/emicklei/go-restful-openapi v1.2.0
	github.com/fatih/color v1.9.0 // indirect
	github.com/gdamore/tcell v1.3.0 // indirect
	github.com/gin-gonic/gin v1.5.0 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.6
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/google/go-github/v31 v31.0.0
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gophercloud/gophercloud v0.8.0 // indirect
	github.com/gosuri/uitable v0.0.4
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.0.3 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/oklog/ulid v1.3.1
	github.com/rivo/tview v0.0.0-20180728193050-6614b16d9037
	github.com/slok/brigadeterm v0.11.1
	github.com/spf13/cobra v1.0.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200219091948-cb0a6d8edb6c // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/gin-gonic/gin.v1 v1.1.5-0.20170702092826-d459835d2b07
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v2.0.0-alpha.0.0.20181016174657-85ed251159e4+incompatible
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe // indirect
)
