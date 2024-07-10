{
  lib,
  pkgs,
}:
pkgs.buildGoModule rec {
  pname = "goji";
  version = "0.1.1";

  src = lib.cleanSource ./.;
  # pkgs.fetchFromGitHub {
  #   owner = "muandane";
  #   repo = "goji";
  #   rev = "v${version}";
  #   sha256 = "sha256-QFll5qr+b+bGl2QJ+rQ72FuETBSeqou/gvcvIY3oDIo=";
  # };

  vendorHash = "sha256-YKnIAviOlLVHaD3lQKhrDlLW1f0cEjY0Az4RyuNWmzg=";

  subPackages = ["."];

  ldflags = [
    "-s"
    "-w"
    "-X goji/cmd.version=${version}"
  ];

  meta = with lib; {
    homepage = "https://github.com/muandane/goji";
    description = " Commitizen-like Emoji Commit Tool written in Go (think cz-emoji and other commitizen adapters but in go) 🚀 ";
    changelog = "https://github.com/muandane/goji/blob/v${version}/CHANGELOG.md";
    license = "Apache 2.0 license Zine El Abidine Moualhi";
    mainProgram = "goji";
  };
}
