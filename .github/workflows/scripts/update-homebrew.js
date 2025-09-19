#!/usr/bin/env node

const { execSync } = require("child_process");
const { readFileSync, writeFileSync, existsSync } = require("fs");
const { join } = require("path");
const { createHash } = require("crypto");

/**
 * Update Homebrew formula with new version and checksums
 * @param {string} version - Version string (e.g., "0.1.0")
 */
async function updateHomebrew(version) {
  console.log(`🍺 Updating Homebrew formula to v${version}...`);

  // Validate version format
  if (!/^\d+\.\d+\.\d+$/.test(version)) {
    console.error("❌ Invalid version format. Expected: x.y.z (e.g., 1.0.0)");
    process.exit(1);
  }

  const projectRoot = process.cwd();
  const homebrewPath = join(projectRoot, "homebrew-tap");
  const formulaPath = join(homebrewPath, "Formula", "occtx.rb");

  // Check if homebrew-tap directory exists
  if (!existsSync(homebrewPath)) {
    console.error("❌ homebrew-tap directory not found. Please clone your homebrew-tap repository first.");
    console.error("   git clone https://github.com/hungthai1401/homebrew-tap.git");
    process.exit(1);
  }

  // Check if formula file exists
  if (!existsSync(formulaPath)) {
    console.error("❌ Formula file not found:", formulaPath);
    process.exit(1);
  }

  // Define platforms and their binary names
  const platforms = [
    { os: "macos", arch: "aarch64", name: "macOS ARM64 (Apple Silicon)" },
    { os: "macos", arch: "x86_64", name: "macOS x86_64 (Intel)" },
    { os: "linux", arch: "aarch64", name: "Linux ARM64" },
    { os: "linux", arch: "x86_64", name: "Linux x86_64" },
  ];

  const checksums = {};

  // Download and calculate checksums for all platforms
  for (const { os, arch, name } of platforms) {
    const binaryName = `occtx-${os}-${arch}`;
    const url = `https://github.com/hungthai1401/occtx/releases/download/v${version}/${binaryName}`;
    console.log(`📥 Calculating checksum for ${name}...`);

    try {
      // Download file temporarily
      const tempFile = `/tmp/${binaryName}`;
      execSync(`curl -L -s -o ${tempFile} "${url}"`, { stdio: 'pipe' });

      // Check if download was successful
      if (!existsSync(tempFile)) {
        throw new Error(`Failed to download ${url}`);
      }

      // Calculate SHA256
      const buffer = readFileSync(tempFile);
      const hash = createHash("sha256").update(buffer).digest("hex");
      checksums[`${os}-${arch}`] = hash;

      // Clean up
      execSync(`rm -f ${tempFile}`);
      console.log(`✅ ${name}: ${hash}`);
    } catch (error) {
      console.error(`❌ Failed to download ${name}:`, error.message);
      console.error(`   URL: ${url}`);
      process.exit(1);
    }
  }

  // Read current formula
  console.log("📝 Reading current formula...");
  const formula = readFileSync(formulaPath, "utf8");

  // Update formula content
  console.log("🔄 Updating formula content...");
  let updatedFormula = formula;

  // Update version
  updatedFormula = updatedFormula.replace(
    /version "[^"]+"/g, 
    `version "${version}"`
  );

  // Update URLs for each platform
  updatedFormula = updatedFormula.replace(
    /url "https:\/\/github\.com\/hungthai1401\/occtx\/releases\/download\/v[^\/]+\/occtx-macos-aarch64"/g,
    `url "https://github.com/hungthai1401/occtx/releases/download/v${version}/occtx-macos-aarch64"`
  );

  updatedFormula = updatedFormula.replace(
    /url "https:\/\/github\.com\/hungthai1401\/occtx\/releases\/download\/v[^\/]+\/occtx-macos-x86_64"/g,
    `url "https://github.com/hungthai1401/occtx/releases/download/v${version}/occtx-macos-x86_64"`
  );

  updatedFormula = updatedFormula.replace(
    /url "https:\/\/github\.com\/hungthai1401\/occtx\/releases\/download\/v[^\/]+\/occtx-linux-aarch64"/g,
    `url "https://github.com/hungthai1401/occtx/releases/download/v${version}/occtx-linux-aarch64"`
  );

  updatedFormula = updatedFormula.replace(
    /url "https:\/\/github\.com\/hungthai1401\/occtx\/releases\/download\/v[^\/]+\/occtx-linux-x86_64"/g,
    `url "https://github.com/hungthai1401/occtx/releases/download/v${version}/occtx-linux-x86_64"`
  );

  // Update SHA256 checksums using a more precise approach
  console.log("🔐 Updating SHA256 checksums...");
  
  // Split content into lines for precise replacement
  const lines = updatedFormula.split('\n');
  let inMacosSection = false;
  let inLinuxSection = false;
  let inArmBlock = false;
  
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    
    // Track which section we're in
    if (line.includes('on_macos do')) {
      inMacosSection = true;
      inLinuxSection = false;
    } else if (line.includes('on_linux do')) {
      inLinuxSection = true;
      inMacosSection = false;
    } else if (line.trim() === 'end' && (inMacosSection || inLinuxSection)) {
      // Check if this is the end of the platform section (not just an if block)
      let foundPlatformBlock = false;
      for (let j = Math.max(0, i - 10); j < i; j++) {
        if (lines[j].includes('on_macos do') || lines[j].includes('on_linux do')) {
          foundPlatformBlock = true;
          break;
        }
      }
      if (foundPlatformBlock && !line.includes('    ')) { // Top-level end
        inMacosSection = false;
        inLinuxSection = false;
      }
    }
    
    // Track ARM vs x86 blocks
    if (line.includes('if Hardware::CPU.arm?')) {
      inArmBlock = true;
    } else if (line.includes('else') && (inMacosSection || inLinuxSection)) {
      inArmBlock = false;
    }
    
    // Update SHA256 hashes
    if (line.includes('sha256') && line.includes('"')) {
      if (inMacosSection && inArmBlock) {
        lines[i] = line.replace(/sha256 "[^"]*"/, `sha256 "${checksums['macos-aarch64']}"`);
      } else if (inMacosSection && !inArmBlock) {
        lines[i] = line.replace(/sha256 "[^"]*"/, `sha256 "${checksums['macos-x86_64']}"`);
      } else if (inLinuxSection && inArmBlock) {
        lines[i] = line.replace(/sha256 "[^"]*"/, `sha256 "${checksums['linux-aarch64']}"`);
      } else if (inLinuxSection && !inArmBlock) {
        lines[i] = line.replace(/sha256 "[^"]*"/, `sha256 "${checksums['linux-x86_64']}"`);
      }
    }
  }
  
  updatedFormula = lines.join('\n');

  // Write updated formula
  writeFileSync(formulaPath, updatedFormula);
  console.log("✅ Updated Formula/occtx.rb");

  // Show the updated content
  console.log("\n📄 Updated formula content:");
  console.log("=" .repeat(50));
  console.log(updatedFormula);
  console.log("=" .repeat(50));

  // Commit and push to homebrew-tap
  console.log("\n🚀 Committing and pushing changes...");
  try {
    // Configure git if needed
    try {
      execSync('git config user.name', { cwd: homebrewPath, stdio: 'pipe' });
    } catch {
      execSync('git config user.name "GitHub Actions"', { cwd: homebrewPath });
      execSync('git config user.email "actions@github.com"', { cwd: homebrewPath });
    }

    // Add changes
    execSync("git add .", { cwd: homebrewPath });
    
    // Check if there are changes to commit
    try {
      execSync("git diff --cached --exit-code", { cwd: homebrewPath, stdio: 'pipe' });
      console.log("ℹ️  No changes to commit");
      return;
    } catch {
      // There are changes, continue with commit
    }

    // Commit with detailed message
    const commitMessage = `Update occtx to version ${version}

- Version: ${version}
- macOS ARM64: ${checksums['macos-aarch64']}
- macOS x86_64: ${checksums['macos-x86_64']}  
- Linux ARM64: ${checksums['linux-aarch64']}
- Linux x86_64: ${checksums['linux-x86_64']}
- Release: https://github.com/hungthai1401/occtx/releases/tag/v${version}

Auto-updated by GitHub Actions`;

    execSync(`git commit -m "${commitMessage}"`, { cwd: homebrewPath });
    // Push using the token if available (for GitHub Actions)
    const token = process.env.GITHUB_TOKEN;
    if (token) {
      execSync(`git remote set-url origin https://${token}@github.com/hungthai1401/homebrew-tap.git`, { cwd: homebrewPath });
    }
    execSync("git push origin main", { cwd: homebrewPath });
    console.log("✅ Pushed to homebrew-tap repository");
  } catch (error) {
    console.error("❌ Failed to update homebrew-tap:", error.message);
    process.exit(1);
  }

  console.log(`\n🍺 Homebrew formula updated to v${version} successfully! 🎉`);
  
  // Summary
  console.log("\n📊 Summary:");
  console.log(`Version: ${version}`);
  console.log(`macOS ARM64 SHA: ${checksums['macos-aarch64']}`);
  console.log(`macOS x86_64 SHA: ${checksums['macos-x86_64']}`);
  console.log(`Linux ARM64 SHA: ${checksums['linux-aarch64']}`);
  console.log(`Linux x86_64 SHA: ${checksums['linux-x86_64']}`);
}

// Export for use in other scripts
module.exports = { updateHomebrew };

// Allow direct execution
if (require.main === module) {
  const version = process.argv[2] || process.env.OCCTX_VERSION;
  if (!version) {
    console.error("❌ Version is required!");
    console.error("");
    console.error("Usage:");
    console.error("  node .github/workflows/scripts/update-homebrew.js <version>");
    console.error("  OCCTX_VERSION=1.0.0 node .github/workflows/scripts/update-homebrew.js");
    console.error("");
    console.error("Examples:");
    console.error("  node .github/workflows/scripts/update-homebrew.js 1.0.0");
    console.error("  OCCTX_VERSION=1.0.0 node .github/workflows/scripts/update-homebrew.js");
    process.exit(1);
  }
  
  updateHomebrew(version).catch((error) => {
    console.error("❌ Script failed:", error.message);
    process.exit(1);
  });
}
