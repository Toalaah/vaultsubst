{pkgs ? (import <nixpkgs> {})}:
with pkgs;
  mkShell {
    buildInputs = [
      go
      golangci-lint
      vault
    ];
  }
