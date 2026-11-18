class Ansimate < Formula
  desc "Take ANSI streams and transform them into compact WebP animations"
  homepage "https://github.com/bengarrett/ansimate"
  url "https://github.com/bengarrett/ansimate/archive/refs/tags/v1.1.5.tar.gz"
  sha256 "b4acae76ff301175d61c3a4d25758cc4fdc659a49ecb7f858abc2c5093d97631"
  version "1.1.5"
  license "MIT"

  livecheck do
    url :stable
    strategy :github_latest
  end

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.version=#{version} -X main.date=#{time.iso8601}"
    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/ansimate"
  end

  test do
    assert_match "ansimate", shell_output("#{bin}/ansimate --version")
  end
end
