package link_analysis

import "yueling_tg/core/plugin"

func Plugins() []plugin.Plugin {

	return []plugin.Plugin{
		NewBili(),
	}
}
