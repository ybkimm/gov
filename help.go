// THIS FILE IS GENERATED!
// DO NOT EDIT
package main

const helpMessage = `Gov is the GO Version manager.

It makes easy to install and uninstall Go versions.
But nothing touches anything else, GOPATH, GOROOT, or the Go binaries that already installed.

Usage:

	gov install (version)    # Installs a specific Go version.
	                         # The directories that installed are as follows:
	                         #     Linux:   ~/.goversions/[version]/
	                         #     Windows: %APPDATA%\.goversions\[version]\
	                         # ... or $GOVERSIONS_PATH(%GOVERSIONS_PATH% on Windows).
	gov uninstall (version)  # Uninstall version.
	gov ls                   # List the installed versions.
	gov (version) [command]  # Run the command with specific version of Go binary.
	gov help                 # Print this message
`