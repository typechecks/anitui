{
  description = "a tui for browsing and streaming anime";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "anitui";
          version = "0.1.0"; # Update this or use a helper to get it from git
          src = ./.;
          vendorHash = null; # Set this after first build or use proxy
          subPackages = [ "cmd/anitui" ];
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ go_1_25 gopls gotools ];
        };
      }
    );
}
