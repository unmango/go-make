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

          # Pages of the GNU Make manual that internal/conformance compares the
          # enumerated syntax against. A hash mismatch means the manual changed,
          # update the hash and review the resulting fixture diff.
          manual = {
            quickRef = pkgs.fetchurl {
              url = "https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html";
              hash = "sha256-0c/jVkSV8ZZo2Xv33f6NSEFfRBp2D1mIKv8lMILwqDQ=";
            };
            specialTargets = pkgs.fetchurl {
              url = "https://www.gnu.org/software/make/manual/html_node/Special-Targets.html";
              hash = "sha256-TTbD1INeiFAyViRpFxCRCECHX5//4pWvcldhyTk0g90=";
            };
          };

          syncQuickRef = pkgs.writeShellApplication {
            name = "sync-quickref";
            runtimeInputs = [ pkgs.go ];
            text = ''
              exec go run ./internal/conformance/cmd/syncquickref \
                ${manual.quickRef} ${manual.specialTargets}
            '';
          };

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

          apps.sync-quickref = {
            type = "app";
            program = pkgs.lib.getExe syncQuickRef;
          };

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
