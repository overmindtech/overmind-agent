module github.com/overmindtech/overmind-agent

go 1.17

// Direct dependencies
require (
	github.com/cakturk/go-netstat v0.0.0-20200220111822-e5b49efee7a5
	github.com/coreos/go-systemd/v22 v22.0.0-20211213101732-f5a75de5182a // Awaiting > 22.3.2
	github.com/elastic/go-sysinfo v1.7.1
	github.com/overmindtech/discovery v0.9.1
	github.com/overmindtech/sdp-go v0.6.0
	github.com/shirou/gopsutil v3.21.10+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	gopkg.in/yaml.v2 v2.4.0
)

// Transitive dependencies
require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/nats-io/jwt/v2 v2.2.0 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20211122170419-d7c1d78a50fc // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/overmindtech/sdpcache v0.1.4 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	golang.org/x/crypto v0.0.0-20211202192323-5770296d904e // indirect
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.64.0 // indirect
	howett.net/plist v0.0.0-20201203080718-1454fab16a06 // indirect
)

require github.com/google/uuid v1.3.0 // indirect
