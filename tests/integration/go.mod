module github.com/mdmourao/go-d1/tests/integration

go 1.26.1

require (
	github.com/joho/godotenv v1.5.1
	github.com/mdmourao/go-d1 v0.0.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/buger/jsonparser v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/mdmourao/go-d1 => ../..
