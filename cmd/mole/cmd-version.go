package main

func init() {
	commands["version"] = command{versionCommand, msgVersionShort}
}

func versionCommand(args []string) error {
	printVersion()
	return nil
}
