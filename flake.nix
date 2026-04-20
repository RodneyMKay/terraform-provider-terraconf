{
    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    };

    outputs = { self, nixpkgs, ... }: let
        supportedSystems = [ "x86_64-linux" "aarch64-darwin" "x86_64-darwin" ];
        forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in {
        devShells = forAllSystems (system: let
            pkgs = import nixpkgs {
                inherit system;
                config = {
                    allowUnfree = true; # Terraform is no longer FOSS :(
                };
            };
        in {
            default = pkgs.mkShell {
                name = "Terraconf Provider Development";

                packages = with pkgs; [
                    # Terraform
                    terraform 
                    terraform-docs

                    # Go
                    go
                    gopls # Language Server
                    golangci-lint # Linter
                ];
            };
        });
    };
}
