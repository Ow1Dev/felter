{
  description = "Dev shell with Go and Bun via flake-parts";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs = { self, ... }@inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem = { pkgs, ... }: {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            bun
            nodejs_20
            corepack
            angular-language-server
            docker
            git
            openssl
            gcc
            jq
            nixd
            golangci-lint
            gofmt
          ];
        };
      };
    };
  } 
