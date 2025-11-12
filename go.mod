module github.com/InjectiveLabs/suplog

go 1.24.0

require (
	github.com/aws/aws-sdk-go v1.25.16
	github.com/bugsnag/bugsnag-go v1.5.3
	github.com/oklog/ulid v1.3.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/bugsnag/panicwrap v1.3.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/bugsnag/bugsnag-go => ./hooks/bugsnag/bugsnag-go
