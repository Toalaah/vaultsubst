{pkgs ? (import <nixpkgs> { config.allowUnfree = true; })}:
with pkgs;
  mkShell {
    buildInputs = [
      go_1_24
      golangci-lint
      vault
      osv-scanner
      goreleaser
    ];
  }
