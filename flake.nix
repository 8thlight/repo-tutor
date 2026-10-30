{
  description = "repo-tutor development environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "aarch64-darwin" "x86_64-darwin" "aarch64-linux" "x86_64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: {
        go = pkgs.go;
        default = pkgs.buildGoModule {
          pname = "repo-tutor";
          version = self.shortRev or self.dirtyShortRev or "dev";
          src = self;
          vendorHash = "sha256-DGFtp74ANqE99+u+f1/N1Nj853HfYdwX94pDxtgG4Ko=";
          subPackages = [ "binary" ];
          env.CGO_ENABLED = 0;
          doCheck = false;
          ldflags = [ "-X repo-tutor/internal/about.version=${self.shortRev or self.dirtyShortRev or "dev"}" ];
          postInstall = "mv $out/bin/binary $out/bin/repo-tutor";
          meta.mainProgram = "repo-tutor";
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          GODEBUG = "gcshrinkstackoff=1";
          packages = [
            pkgs.go
            pkgs.gopls
            pkgs.just
            pkgs.gh
            pkgs.lychee
            pkgs.markdownlint-cli2
            pkgs.vale
            (pkgs.python3.withPackages (ps: [
              ps.mdformat
              ps.mdformat-gfm
              ps.mdformat-frontmatter
            ]))
          ];
        };
      });
    };
}
