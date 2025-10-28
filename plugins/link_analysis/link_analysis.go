package link_analysis

import "yueling_tg/pkg/plugin"

func Plugins() []plugin.Plugin {

	return []plugin.Plugin{
		NewBili(),
	}
}
