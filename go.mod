module e-commerce-app

go 1.18

require (
	github.com/cloudevents/sdk-go/v2 v2.12.0
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/lib/pq v1.10.6
	github.com/stretchr/testify v1.8.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace aws-step-functions-long-lived-transactions/models => ../models
