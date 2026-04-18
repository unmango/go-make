{
  description = "Makefile parsing and utilities in Go";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";
    flake-parts.url = "github:hercules-ci/flake-parts";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;

      imports = [ inputs.treefmt-nix.flakeModule ];

      perSystem =
        {
          inputs',
          pkgs,
          system,
          ...
        }:
        let
          inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication mkGoEnv;

          goEnv = mkGoEnv { pwd = ./.; };

          goMake = buildGoApplication {
            pname = "go-make";
            version = "0.8.0";
            src = ./.;

            modules = ./gomod2nix.toml;

            nativeBuildInputs = [ pkgs.ginkgo ];
          };
        in
        {
          _module.args.pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ inputs.gomod2nix.overlays.default ];
          };

          packages.goMake = goMake;
          packages.default = goMake;

          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              bash # For copilot
              gnumake
              go
              goEnv
              gomod2nix
              nixfmt
            ];
          };

          treefmt = {
            programs.nixfmt.enable = true;
          };
        };
    };
}
