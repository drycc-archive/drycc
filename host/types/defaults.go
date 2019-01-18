// +build linux freebsd

package host

var (
        Wildcard = int64(-1)
        // DefaultCapabilities is the default list of capabilities which are set inside
        // a container, taken from:
        // https://github.com/opencontainers/runc/blob/v1.0.0-rc1/libcontainer/SPEC.md#security
        DefaultCapabilities = []string{
                "CAP_NET_RAW",
                "CAP_NET_BIND_SERVICE",
                "CAP_DAC_OVERRIDE",
                "CAP_SETFCAP",
                "CAP_SETPCAP",
                "CAP_SETGID",
                       "CAP_SETUID",
                       "CAP_MKNOD",
                "CAP_CHOWN",
                "CAP_FOWNER",
                "CAP_FSETID",
                "CAP_KILL",
                "CAP_SYS_CHROOT",
        }
        // DefaultSimpleDevices are devices that are to be both allowed and created.
        DefaultSimpleDevices = []*Device{
                // /dev/null and zero
                {
                        Path:        "/dev/null",
                        Type:        'c',
                        Major:       1,
                        Minor:       3,
                        Permissions: "rwm",
                        FileMode:    0666,
                },
                {
                        Path:        "/dev/zero",
                        Type:        'c',
                        Major:       1,
                        Minor:       5,
                        Permissions: "rwm",
                        FileMode:    0666,
                },

                {
                        Path:        "/dev/full",
                        Type:        'c',
                        Major:       1,
                        Minor:       7,
                        Permissions: "rwm",
                        FileMode:    0666,
                },

                // consoles and ttys
                {
                        Path:        "/dev/tty",
                        Type:        'c',
                        Major:       5,
                        Minor:       0,
                        Permissions: "rwm",
                        FileMode:    0666,
                },

                // /dev/urandom,/dev/random
                {
                        Path:        "/dev/urandom",
                        Type:        'c',
                        Major:       1,
                        Minor:       9,
                        Permissions: "rwm",
                        FileMode:    0666,
                },
                {
                        Path:        "/dev/random",
                        Type:        'c',
                        Major:       1,
                        Minor:       8,
                        Permissions: "rwm",
                        FileMode:    0666,
                },
        }
        DefaultAllowedDevices = append([]*Device{
                // allow mknod for any device
                {
                        Type:        'c',
                        Major:       Wildcard,
                        Minor:       Wildcard,
                        Permissions: "m",
                },
                {
                        Type:        'b',
                        Major:       Wildcard,
                        Minor:       Wildcard,
                        Permissions: "m",
                },

                {
                        Path:        "/dev/console",
                        Type:        'c',
                        Major:       5,
                        Minor:       1,
                        Permissions: "rwm",
                },
                // /dev/pts/ - pts namespaces are "coming soon"
                {
                        Path:        "",
                        Type:        'c',
                        Major:       136,
                        Minor:       Wildcard,
                        Permissions: "rwm",
                },
                {
                        Path:        "",
                        Type:        'c',
                        Major:       5,
                        Minor:       2,
                        Permissions: "rwm",
                },

                // tuntap
                {
                        Path:        "",
                        Type:        'c',
                        Major:       10,
                        Minor:       200,
                        Permissions: "rwm",
                },
        }, DefaultSimpleDevices...)
        DefaultAutoCreatedDevices = append([]*Device{}, DefaultSimpleDevices...)
)
