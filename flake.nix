{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.utils.follows = "utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      utils,
      gomod2nix,
      ...
    }@inputs:

    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ gomod2nix.overlays.default ];
      };
    in
    {
      devShells.${system}.default =
        let
          goEnv = pkgs.mkGoEnv { pwd = ./.; };
        in
        pkgs.mkShell {
          buildInputs = with pkgs; [
            goEnv
            go
            gopls
            gotools
            go-tools
            golangci-lint
            gomod2nix.packages.${system}.default
            (python312.withPackages (
              p: with p; [
                requests
                lxml
              ]
            ))
          ];
        };
      formatter.x86_64-linux = nixpkgs.legacyPackages.x86_64-linux.nixfmt-rfc-style;

    };
}
