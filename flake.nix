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
        version = builtins.substring 0 10 self.lastModifiedDate + "-" + (self.shortRev or "dirty");
      in
      {
        packages.ssm = pkgs.buildGoModule {
          pname = "ssm";
          inherit version;
          src = self;
          vendorHash = "sha256-FKQAgmNId+9Pv+gc7QhYT0zu16/4UvEfycQClhlEQWc=";
          ldflags = [
            "-s" "-w"
            "-X main.BuildVersion=${version}"
            "-X main.BuildDate=${self.lastModifiedDate}"
            "-X main.BuildSHA=${self.shortRev or "dirty"}"
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
