class Ansimate < Formula
  desc "Take ANSI streams and transform them into compact WebP animations"
  homepage "https://github.com/bengarrett/ansimate"
  url "https://github.com/bengarrett/ansimate/archive/refs/tags/v1.2.0.tar.gz"
  sha256 "50e15f3371ddf10c1f5e9f493c126d316a7dc3185e6b6f25ab727ef6eb7d8fbf"
  version "1.2.0"
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
