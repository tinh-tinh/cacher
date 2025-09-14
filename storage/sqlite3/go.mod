module github.com/tinh-tinh/cacher/storage/sqlite3

go 1.22.2

require (
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/stretchr/testify v1.9.0
	github.com/tinh-tinh/cacher/v2 v2.4.0
	github.com/tinh-tinh/tinhtinh/v2 v2.3.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/tinh-tinh/cacher => ../../
