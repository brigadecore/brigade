module github.com/brigadecore/brigade/v2

go 1.15

replace github.com/brigadecore/brigade/sdk/v2 => ../sdk/v2

require (
	github.com/AlecAivazis/survey/v2 v2.0.7
	github.com/Azure/go-amqp v0.13.1
	github.com/bacongobbler/browser v1.1.0
	github.com/brigadecore/brigade/sdk/v2 v2.0.0-20200923171232-9f56c474d8bf
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/fatih/color v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-git/go-billy v4.2.0+incompatible
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/go-git/go-git v4.7.0+incompatible
	github.com/go-git/go-git/v5 v5.2.0
	github.com/gorilla/mux v1.7.4
	github.com/gosuri/uitable v0.0.4
	github.com/kr/pretty v0.2.1 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20200921180117-858c6e7e6b7e // indirect
	github.com/rs/cors v1.7.0
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.2.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.mongodb.org/mongo-driver v1.4.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
)
