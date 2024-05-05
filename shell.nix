{pkgs ? (import <nixpkgs> { config.allowUnfree = true; })}:
with pkgs;
  mkShell {
    buildInputs = [
      go_1_22
      golangci-lint
      vault
      osv-scanner
    ];
  }
