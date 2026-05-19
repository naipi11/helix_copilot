class HelixCopilot < Formula
  desc "Helix GitHub Copilot bridge and patched ghost text workflow"
  homepage "https://github.com/naipi11/helix_copilot"
  version "0.1.0"
  license "MIT"

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/naipi11/helix_copilot/releases/download/v0.1.0/helix-copilot_linux_arm64.tar.gz"
      sha256 "TODO"
    else
      url "https://github.com/naipi11/helix_copilot/releases/download/v0.1.0/helix-copilot_linux_amd64.tar.gz"
      sha256 "TODO"
    end
  end

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/naipi11/helix_copilot/releases/download/v0.1.0/helix-copilot_darwin_arm64.tar.gz"
      sha256 "TODO"
    else
      url "https://github.com/naipi11/helix_copilot/releases/download/v0.1.0/helix-copilot_darwin_amd64.tar.gz"
      sha256 "TODO"
    end
  end

  depends_on "node"

  def install
    bin.install "helix-copilot"
    bin.install "hx" if File.exist?("hx")
  end

  test do
    system "#{bin}/helix-copilot", "--help"
  end
end
