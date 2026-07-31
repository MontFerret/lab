package runtime

func normalizeBinaryRuntimeProtocol(protocol Protocol) (Protocol, error) {
	if protocol == ProtocolUnknown {
		return ProtocolCLIDirect, nil
	}

	return ParseProtocol(string(protocol))
}

func validateCLIDirectProtocol(opts BinaryOptions) error {
	if opts.FSPolicy.hasSettings() {
		return ProtocolCLIDirect.unsupported("filesystem policy options")
	}

	if opts.HTTPPolicy.hasSettings() {
		return ProtocolCLIDirect.unsupported("HTTP policy options")
	}

	return nil
}
