{ pkgs, self }:

pkgs.buildGoModule {
  pname = "beatport-dl";
  version = self.shortRev or "dirty";
  src = ../.;
  vendorHash = null;

  # The main package is in cmd/beatport-dl
  subPackages = [ "cmd/beatport-dl" ];

  # Enable CGo for taglib integration
  env.CGO_ENABLED = 1;

  # Native dependencies required for the build process itself
  nativeBuildInputs = [
    pkgs.zig
  ];

  # Libraries to link against
  buildInputs = [
    pkgs.taglib
    pkgs.zlib
  ];

  ldflags = [ "-w" "-linkmode=external" "-extldflags=-lstdc++" ];

}
