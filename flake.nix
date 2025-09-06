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
      in {
        overlays.default = final: prev: {
          beatport-dl = final.callPackage ./nix/package.nix { inherit pkgs self; };
        };

        packages = {
          beatport-dl = pkgs.callPackage ./nix/package.nix { inherit pkgs self; };
          default = self.packages."${system}".beatport-dl;
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ self.packages."${system}".beatport-dl ];
          packages = [ pkgs.go pkgs.gopls ];
        };
      }
    );
}
