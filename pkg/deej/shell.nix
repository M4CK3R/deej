{ pkgs ? import <nixpkgs> { } }:
let
buildInputs=with pkgs; [
    go
    pkg-config
    gtk4
    libayatana-indicator
    libindicator
  ];
in
pkgs.mkShell {
  buildInputs = buildInputs;

  LD_LIBRARY_PATH = "$LD_LIBRARY_PATH:${pkgs.lib.makeLibraryPath buildInputs}";
}
