#!/bin/sh
# Render the brew formula and scoop manifest from dist/ binaries; requires
# an exact release tag so URLs, stamps, and hashes agree.
set -e
VERSION="$(git describe --tags --exact-match)"
BARE="${VERSION#v}"
BASE="https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/releases/download/${VERSION}"

sha() { shasum -a 256 "$1" | cut -d' ' -f1; }
DARWIN_ARM64="$(sha dist/agent-compose-darwin-arm64)"
LINUX_AMD64="$(sha dist/agent-compose-linux-amd64)"
LINUX_ARM64="$(sha dist/agent-compose-linux-arm64)"
WINDOWS_AMD64="$(sha dist/agent-compose-windows-amd64.exe)"

cat > dist/agent-compose.rb <<EOF
class AgentCompose < Formula
  desc "Kai's personality engine for agent context"
  homepage "https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose"
  version "${BARE}"
  license "MIT"

  on_macos do
    on_arm do
      url "${BASE}/agent-compose-darwin-arm64"
      sha256 "${DARWIN_ARM64}"
    end
  end
  on_linux do
    on_intel do
      url "${BASE}/agent-compose-linux-amd64"
      sha256 "${LINUX_AMD64}"
    end
    on_arm do
      url "${BASE}/agent-compose-linux-arm64"
      sha256 "${LINUX_ARM64}"
    end
  end

  def install
    bin.install Dir["agent-compose-*"].first => "agent-compose"
    bin.install_symlink "agent-compose" => "acompose"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/agent-compose version")
  end
end
EOF

cat > dist/agent-compose.json <<EOF
{
    "version": "${BARE}",
    "description": "Kai's personality engine for agent context",
    "homepage": "https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose",
    "license": "MIT",
    "architecture": {
        "64bit": {
            "url": "${BASE}/agent-compose-windows-amd64.exe",
            "hash": "${WINDOWS_AMD64}",
            "bin": [
                ["agent-compose-windows-amd64.exe", "agent-compose"],
                ["agent-compose-windows-amd64.exe", "acompose", "compose"]
            ]
        }
    }
}
EOF

echo "dist/agent-compose.rb"
echo "dist/agent-compose.json"
