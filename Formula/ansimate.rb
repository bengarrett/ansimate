class Ansimate < Formula
  desc "Take ANSI streams and transform them into compact WebP animations"
  homepage "https://github.com/bengarrett/ansimate"
  url "https://github.com/bengarrett/ansimate/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "107af0f32ed4c4ed2312ca5d7926d97ae667c3f27ec8be8244000684fd61f72d"
  version "1.0.0"
  license "MIT"

  livecheck do
    url :stable
    strategy :github_latest
  end

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/ansimate"
  end

  test do
    assert_match "ansimate", shell_output("#{bin}/ansimate --version")
  end
end
