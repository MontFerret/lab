package runtime

import "errors"

func validateBinaryRuntimeProtocol(protocol Protocol) error {
	switch protocol {
	case ProtocolCLI, ProtocolCLIDirect:
		return nil
	default:
		return errors.New("unsupported binary runtime protocol: " + string(protocol))
	}
}

func validateCLIDirectProtocol(opts BinaryOptions) error {
	if len(opts.Params) > 0 {
		return ProtocolCLIDirect.unsupported("bind parameters")
	}

	if len(opts.Flags) > 0 {
		return ProtocolCLIDirect.unsupported("runtime flags")
	}

	if opts.FSPolicy.hasSettings() {
		return ProtocolCLIDirect.unsupported("filesystem policy options")
	}

	if opts.HTTPPolicy.hasSettings() {
		return ProtocolCLIDirect.unsupported("HTTP policy options")
	}

	return nil
}
