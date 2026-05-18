{
  description = "SSM — Secure Shell Manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.ssm = pkgs.buildGoModule {
          pname = "ssm";
          version = "2.2.1";
          src = self;
          vendorHash = "sha256-iJVgkKqrDxRASRjEp8SVwFjsLR9l5gsdpFPbewrg9VI=";
          ldflags = [
            "-s" "-w"
            "-X main.BuildVersion=2.2.1"
          ];
          subPackages = [ "." ];
        };
        packages.default = self.packages.${system}.ssm;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_26
            golangci-lint
            goreleaser
          ];
        };
      });
}
