package staticserver

import "github.com/MontFerret/lab/v2/pkg/localserver"

const defaultHost = localserver.DefaultHost

type Settings = localserver.Settings

func ResolveSettings(settings Settings) (Settings, error) {
	return localserver.ResolveSettings(settings)
}
