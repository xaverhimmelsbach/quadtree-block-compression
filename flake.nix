{
  description = "Build a compressed quadtree image from block-based sub-images";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in
      {
        devShell = pkgs.mkShell {
          buildInputs = with pkgs;[
            go
            gotools
            golangci-lint
            gopls
            go-outline
            gopkgs
            delve
          ];
        };
      });
}
