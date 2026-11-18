class Ansimate < Formula
  desc "Take ANSI streams and transform them into compact WebP animations"
  homepage "https://github.com/bengarrett/ansimate"
  url "https://github.com/bengarrett/ansimate/archive/refs/tags/v1.1.0.tar.gz"
  sha256 "90dceaec77a2bc8e9da7c682302ca92c81477a83c8dac51b6a74025b4765fe4d"
  version "1.1.0"
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
