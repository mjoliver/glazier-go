module github.com/mjoliver/glazier-go

go 1.25.3

// Use local copy of upstream glazier libraries.
// When git is available, remove this replace and run:
//   go get github.com/google/glazier@latest
replace github.com/google/glazier => ./third_party/glazier

require (
	github.com/StackExchange/wmi v1.2.1
	github.com/capnspacehook/taskmaster v0.0.0-20210519235353-1629df7c85e9
	github.com/google/glazier v0.0.0-20221201205010-c6e59b1b4ae6
	golang.org/x/sys v0.41.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/cabbie v1.0.5 // indirect
	github.com/google/deck v1.1.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/iamacarpet/go-win64api v0.0.0-20240507095429-873e84e85847 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rickb777/date v1.21.1 // indirect
	github.com/rickb777/plural v1.4.4 // indirect
	github.com/scjalliance/comshim v0.0.0-20251021001035-b69f3cdad6f3 // indirect
	gopkg.in/toast.v1 v1.0.0-20180812000517-0a84660828b2 // indirect
)
