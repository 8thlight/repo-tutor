{
  description = "repo-tutor development environment";

  inputs.threnos.url = "github:8thlight-daveloper/threnos";
  inputs.nixpkgs.follows = "threnos/nixpkgs";

  outputs = { self, threnos, ... }:
    let
      releaseVersion = builtins.replaceStrings [ "\n" ] [ "" ] (builtins.readFile ./VERSION);
    in
    threnos.lib.projectOutputs {
      src = self;
      recipes = [ "check" "mutate" "mutate-plan" "mutate-summary" "mutate-report" ];
      devShell = pkgs: {
        packages = [ pkgs.go pkgs.gopls ];
        env.GODEBUG = "gcshrinkstackoff=1";
        shellHook = "unset GOROOT";
      };
      packages = pkgs: {
        go = pkgs.go;
        default = pkgs.buildGoModule {
          pname = "repo-tutor";
          version = releaseVersion;
          src = self;
          vendorHash = "sha256-DGFtp74ANqE99+u+f1/N1Nj853HfYdwX94pDxtgG4Ko=";
          subPackages = [ "binary" ];
          env.CGO_ENABLED = 0;
          doCheck = false;
          ldflags = [ "-X repo-tutor/internal/about.version=v${releaseVersion}" ];
          postInstall = "mv $out/bin/binary $out/bin/repo-tutor";
          meta.mainProgram = "repo-tutor";
        };
      };
    };
}
