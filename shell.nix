{pkgs ? (import <nixpkgs> {})}:
with pkgs;
  mkShell {
    buildInputs = [
      go_1_22
      golangci-lint
      vault
      osv-scanner
    ];
  }
