package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const helpText = `protoc plugin for Go for the Very Simple RPC protocol
Usage: protoc --go-vsrpc_out=. path/to/service.proto
`

var (
	Version    = "(unknown)"
	Commit     = ""
	CommitDate = ""
	TreeState  = ""
)

func versionText() string {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	fmt.Fprintf(buf, "%s\n", Version)
	if Commit != "" {
		fmt.Fprintf(buf, "Commit=%s\n", Commit)
	}
	if CommitDate != "" {
		fmt.Fprintf(buf, "CommitDate=%s\n", CommitDate)
	}
	if TreeState != "" {
		fmt.Fprintf(buf, "TreeState=%s\n", TreeState)
	}
	return buf.String()
}

func main() {
	args := os.Args
	argsLen := uint(len(args))
	if argsLen > 0 {
		args = args[1:]
		argsLen--
	}

	wantHelp := false
	wantVersion := false
	for i := uint(0); i < argsLen; i++ {
		switch args[i] {
		case "/?":
			fallthrough
		case "-?":
			fallthrough
		case "-h":
			fallthrough
		case "--help":
			wantHelp = true

		case "-v":
			fallthrough
		case "-V":
			fallthrough
		case "--version":
			wantVersion = true
		}
	}

	if wantVersion {
		io.WriteString(os.Stdout, versionText())
		return
	}

	if wantHelp {
		io.WriteString(os.Stdout, helpText)
		return
	}

	o := protogen.Options{
		ParamFunc: func(name string, value string) error {
			return fmt.Errorf("unknown parameter name %q", name)
		},
	}
	o.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, file := range plugin.Files {
			if file.Generate {
				g := &Generator{Plugin: plugin, File: file}
				g.GenerateFile()
			}
		}
		return nil
	})
}
