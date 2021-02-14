package config

import (
	"flag"
	"fmt"
	"os/user"
	"path"
)

// File name in the user's home directory where we store the settings.
const defaultProfile = ".moprc"

// Flags represents the expected available command line flags.
type Flags struct {
	Profile string
	Version bool
}

// ParseFlags parses flags from user input and applies defaults where applicable.
func ParseFlags() *Flags {
	const profileUsage = "the path to the configuration file"
	const versionUsage = "provide the version information"

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	defaultProfilePath := path.Join(usr.HomeDir, defaultProfile)

	flags := &Flags{}
	flag.StringVar(&flags.Profile, "profile", defaultProfilePath, profileUsage)
	flag.StringVar(&flags.Profile, "p", defaultProfilePath, fmt.Sprintf("%s (shorthand)", profileUsage))
	flag.BoolVar(&flags.Version, "version", false, versionUsage)
	flag.Parse()

	return flags
}
