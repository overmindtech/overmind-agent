module github.com/overmindtech/overmind-agent

go 1.17

// This is required as we are still waiting for a release to be made after PR
// #385 was merged. Once this has been done this can be removed. The newly
// release version will be >= 22.3.3
//
// Ref: https://github.com/coreos/go-systemd/pull/385
replace github.com/coreos/go-systemd/v22 => github.com/coreos/go-systemd/v22 v22.0.0-20211213101732-f5a75de5182a

// Direct dependencies
require (
	github.com/cakturk/go-netstat v0.0.0-20200220111822-e5b49efee7a5
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/elastic/go-sysinfo v1.7.1
	github.com/overmindtech/discovery v0.11.2
	github.com/overmindtech/sdp-go v0.6.1
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	gopkg.in/yaml.v2 v2.4.0
)

// Transitive dependencies
require (
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/nats-io/nats-server/v2 v2.7.3 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20220121202836-972a071d373d // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/overmindtech/sdpcache v0.3.1 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/afero v1.8.1 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	howett.net/plist v1.0.0 // indirect
)
