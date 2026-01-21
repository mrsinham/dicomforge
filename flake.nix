{
  description = "CLI tool to generate valid DICOM MRI series for testing";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "dicomforge";
            version = "1.0.3";

            src = ./.;

            vendorHash = "sha256-EltQlVxP5UY1V1wMfDu07/EOtHAfIdeOss+ocpndZWs=";

            ldflags = [
              "-s"
              "-w"
              "-X main.version=1.0.3"
            ];

            meta = with pkgs.lib; {
              description = "CLI tool to generate valid DICOM MRI series for testing medical imaging platforms";
              homepage = "https://github.com/mrsinham/dicomforge";
              license = licenses.mit;
              maintainers = [ ];
              mainProgram = "dicomforge";
            };
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_24
            gopls
            gotools
            go-tools
          ];
        };
      }
    );
}
