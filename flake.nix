{
  description = "A Nix flake for beatport-dl";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        beatport-dl = pkgs.buildGoModule {
          pname = "beatport-dl";
          version = self.shortRev or "dirty";
          src = ./.;
          vendorHash = null;

          # The main package is in cmd/beatport-dl
          subPackages = [ "cmd/beatport-dl" ];

          # Enable CGo for taglib integration
          env.CGO_ENABLED = 1;

          # Native dependencies required for the build process itself
          nativeBuildInputs = [
            pkgs.zig # For zig cc/c++
          ];

          # Libraries to link against
          buildInputs = [
            pkgs.taglib
            pkgs.zlib
          ];

          ldflags = [ "-w" "-linkmode=external" "-extldflags=-lstdc++" ];

        };

      in {
        packages.default = beatport-dl;

        devShells.default = pkgs.mkShell {
          inputsFrom = [ beatport-dl ];
          packages = [
            pkgs.go
            pkgs.gopls
          ];
        };
      }
    );
}
