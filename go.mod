module github.com/prometheus-community/gaussdb_exporter

go 1.21.5

require (
	config v0.0.0
	gitee.com/opengauss/openGauss-connector-go-pq v1.0.4
	github.com/alecthomas/kingpin/v2 v2.3.2
	github.com/blang/semver/v4 v4.0.0
	github.com/go-kit/log v0.2.1
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/common v0.45.0
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/yaml.v2 v2.4.0
	myencrypt v0.0.0-00010101000000-000000000000
)

require (
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace myencrypt => ./myencrypt

replace config => ./config
