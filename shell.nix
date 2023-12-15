{pkgs ? (import <nixpkgs> {})}:
with pkgs;
  mkShell {
    buildInputs = [
      go_1_21
      golangci-lint
      vault
    ];
  }
