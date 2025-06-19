module github.com/InjectiveLabs/suplog

go 1.15

require (
	github.com/aws/aws-sdk-go v1.25.16
	github.com/bugsnag/bugsnag-go v1.5.3
	github.com/oklog/ulid v1.3.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
)

require golang.org/x/net v0.0.0-20220517181318-183a9ca12b87 // indirect

replace github.com/bugsnag/bugsnag-go => ./hooks/bugsnag/bugsnag-go
