package random

import "yueling_tg/core/plugin"

func Plugins() []plugin.Plugin {

	return []plugin.Plugin{
		NewRollPlugin(),
		NewEmotePlugin(),
	}

}
