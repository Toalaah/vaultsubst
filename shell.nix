{pkgs ? (import <nixpkgs> {})}:
with pkgs;
  mkShell {
    buildInputs = [
      go_1_26
      golangci-lint
      osv-scanner
      goreleaser
    ];
  }
