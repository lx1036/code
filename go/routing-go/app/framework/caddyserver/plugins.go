package caddy

import "sort"

var (
	// serverTypes is a map of registered server types.
	serverTypes = make(map[string]ServerType)
)

// DescribePlugins returns a string describing the registered plugins.
func DescribePlugins() string {
	pl := ListPlugins()

	str := "Server types:\n"
	for _, name := range pl["server_types"] {
		str += "  " + name + "\n"
	}

	str += "\nCaddyfile loaders:\n"
	for _, name := range pl["caddyfile_loaders"] {
		str += "  " + name + "\n"
	}

	if len(pl["event_hooks"]) > 0 {
		str += "\nEvent hook plugins:\n"
		for _, name := range pl["event_hooks"] {
			str += "  hook." + name + "\n"
		}
	}

	if len(pl["clustering"]) > 0 {
		str += "\nClustering plugins:\n"
		for _, name := range pl["clustering"] {
			str += "  " + name + "\n"
		}
	}

	str += "\nOther plugins:\n"
	for _, name := range pl["others"] {
		str += "  " + name + "\n"
	}

	return str
}

// ListPlugins makes a list of the registered plugins,
// keyed by plugin type.
func ListPlugins() map[string][]string {
	p := make(map[string][]string)

	// server type plugins
	for name := range serverTypes {
		p["server_types"] = append(p["server_types"], name)
	}

	// caddyfile loaders in registration order
	for _, loader := range caddyfileLoaders {
		p["caddyfile_loaders"] = append(p["caddyfile_loaders"], loader.name)
	}
	if defaultCaddyfileLoader.name != "" {
		p["caddyfile_loaders"] = append(p["caddyfile_loaders"], defaultCaddyfileLoader.name)
	}

	// List the event hook plugins
	eventHooks.Range(func(k, _ interface{}) bool {
		p["event_hooks"] = append(p["event_hooks"], k.(string))
		return true
	})

	// alphabetize the rest of the plugins
	var others []string
	for stype, stypePlugins := range plugins {
		for name := range stypePlugins {
			var s string
			if stype != "" {
				s = stype + "."
			}
			s += name
			others = append(others, s)
		}
	}

	sort.Strings(others)
	for _, name := range others {
		p["others"] = append(p["others"], name)
	}

	return p
}
