{
  description = "Navidrome lyrics scrape plugin dev environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        version = "0.1.0";
        pkgs = import nixpkgs { inherit system; };
        plugin = pkgs.buildGo125Module {
          pname = "navidrome-shazam-plugin";
          version = version;

          src = ./plugin;

          vendorHash = "sha256-yFPWnIFMB9NeCDu3Km5WwfG3iLpnh/gBUa9dh8+QHFY=";

          nativeBuildInputs = [ pkgs.tinygo ];

          buildPhase = ''
            export HOME=$(mktemp -d)

            tinygo build \
                -target=wasip1 \
                -buildmode=c-shared \
                -opt=z \
                -no-debug \
                -panic=trap \
                -gc=leaking \
                -o plugin.wasm .
          '';

          installPhase = ''
            mkdir -p $out/{bin,share}
            cp plugin.wasm $out/bin

            cp manifest.json $out/share
          '';

          dontCheck = true;
        };
      in
      {
        packages.wasmPlugin = plugin;
        packages.default = pkgs.stdenv.mkDerivation {
          name = "navidrome-shazam-plugin-dev";

          src = plugin;

          buildInputs = with pkgs; [
            zip
            binaryen
            jq
            advancecomp
          ];

          buildPhase = ''
            wasm-opt -Oz \
                --strip-debug \
                --strip-producers \
                --strip-target-features \
                --vacuum \
                --dce \
                --remove-unused-module-elements \
                --converge \
                ${plugin}/bin/plugin.wasm -o plugin.wasm
            jq -c . ${plugin}/share/manifest.json > manifest.json

            touch -t 202001010000.00 manifest.json plugin.wasm
            TZ=UTC zip -9 -X -D out.zip manifest.json plugin.wasm
            advzip -z -4 out.zip
          '';

          installPhase = ''
            mkdir -p $out/bin
            cp out.zip $out/bin/navidrome-shazam-plugin.ndp
          '';
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            tinygo
            gopls
            gofumpt
            zip
            just

            binaryen

            nixfmt-tree
            pinact
          ];
        };
      }
    );
}
