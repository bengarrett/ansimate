class Ansimate < Formula
  desc "Take ANSI streams and transform them into compact WebP animations"
  homepage "https://github.com/bengarrett/ansimate"
  url "https://github.com/bengarrett/ansimate/archive/refs/tags/v1.1.6.tar.gz"
  sha256 "6059ae1005f3d7e0e10c10210d4421bfd327c70fbe7afc941f42ae603353a26e"
  version "1.1.6"
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
