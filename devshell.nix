{ inputs, ... }:
{
  imports = [
    inputs.devshell.flakeModule
    inputs.pre-commit-hooks-nix.flakeModule
  ];

  perSystem =
    { pkgs, config, ... }:
    {
      devshells.default = {
        devshell = {
          name = "GIVC";
          motd = ''
            ❄️ Welcome to the {14}{bold}Welcome to givc-devshell{reset} devshell ❄️
            $(type -p menu &>/dev/null && menu)
            $(type -p update-pre-commit-hooks &>/dev/null && update-pre-commit-hooks)
          '';
        };
        packages = with pkgs; [
          config.treefmt.build.wrapper
          reuse
          go
          gotests
          gopls
          golangci-lint
          rustc
          rustfmt
          cargo
          pkgs.stdenv.cc # Need for build rust components
          protobuf
          protoc-gen-go
          protoc-gen-go-grpc
          grpcurl
        ];
        commands = [
          {
            name = "update-pre-commit-hooks";
            command = config.pre-commit.installationScript;
            category = "tools";
            help = "update git pre-commit hooks";
          }
          {
            help = "Generate go files from protobuffers";
            name = "protogen";
            command = "./api/protoc.sh";
          }
          {
            help = "Like cURL, but for gRPC: Command-line tool for interacting with gRPC servers";
            name = "gcl";
            command = "grpcurl";
          }
          {
            help = "Update go dependencies";
            name = "go-update";
            command = "go get -u ./... && go mod tidy && echo Done - do not forget to update the vendorHash in the packages (use pkgs.lib.fakeHash).";
          }
          {
            help = "golang linter";
            package = "golangci-lint";
            category = "linters";
          }
        ];
      };
      pre-commit.settings = {
        hooks.treefmt.enable = true;
        hooks.treefmt.package = config.treefmt.build.wrapper;
      };
    };
}
